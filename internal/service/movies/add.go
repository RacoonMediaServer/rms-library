package movies

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/RacoonMediaServer/rms-library/v3/internal/lock"
	"github.com/RacoonMediaServer/rms-library/v3/internal/model"
	"github.com/RacoonMediaServer/rms-library/v3/internal/schedule"
	"github.com/RacoonMediaServer/rms-library/v3/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-library/v3/pkg/selector"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

const watchInterval = 1 * time.Minute
const notifyTimeout = 15 * time.Second
const checkReleasesInterval = 24 * time.Hour

var errAnyTorrentsNotFound = errors.New("any torrents not found")

func (l MoviesService) Add(ctx context.Context, id model.ID, list rms_library.List) error {
	info, err := l.db.GetMovieInfo(ctx, id)
	if err != nil {
		return fmt.Errorf("movie '%s' not found in cache: %w", id.String(), err)
	}

	mov := model.Movie{
		ListItem: model.ListItem{
			ID:          id,
			CreatedAt:   time.Now(),
			Title:       info.Title,
			List:        list,
			ContentType: rms_library.ContentType_TypeMovies,
			Category:    model.GetVideoCategory(info.Type),
		},
		Info: *info,
	}

	if err = l.db.AddMovie(ctx, &mov); err != nil {
		return fmt.Errorf("add movie to database failed: %s", err)
	}

	task := schedule.Task{
		Group: id.String(),
		Fn: schedule.GetRetryWrapper(
			logger.Fields(map[string]interface{}{
				"op":    "downloadMovieContent",
				"id":    id.String(),
				"title": mov.Info.Title,
			}),
			func(log logger.Logger, ctx context.Context) error {
				return l.asyncDownloadContent(log, ctx, id)
			},
		),
	}
	l.sched.Add(&task)

	return nil
}

func (l MoviesService) downloadContent(log logger.Logger, ctx context.Context, mov *model.Movie) error {
	if mov.List == rms_library.List_Archive {
		if err := l.searchAndSave(log, ctx, mov); err != nil {
			return fmt.Errorf("search and save content failed: %w", err)
		}
	} else {
		if err := l.searchAndDownload(log, ctx, mov); err != nil {
			return fmt.Errorf("search and download content failed: %w", err)
		}
	}

	return nil
}

func (l MoviesService) asyncDownloadContent(log logger.Logger, ctx context.Context, id model.ID) error {
	lk, err := lock.TimedLock(ctx, l.lk, id, lockWait)
	if err != nil {
		return fmt.Errorf("Lock item failed: %w", err)
	}
	defer lk.Unlock()

	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return fmt.Errorf("load movie from database failed: %w", err)
	}

	if err = l.downloadContent(log, ctx, mov); err != nil {
		return fmt.Errorf("add content failed: %w", err)
	}

	l.startWatchers(mov)
	return nil
}

func (l MoviesService) searchAndDownload(log logger.Logger, ctx context.Context, mov *model.Movie) error {
	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)

	var strategy movsearch.Strategy
	sel := l.getMovieSelector(mov)

	if mov.Info.Type == rms_library.MovieType_TvSeries {
		strategy = &movsearch.FullStrategy{Engine: searchEngine, Selector: sel}
	} else {
		strategy = &movsearch.SimpleStrategy{Engine: searchEngine, Selector: sel}
	}

	selopts := selector.Options{
		Criteria:  selector.CriteriaQuality,
		MediaType: media.Movies,
		Query:     mov.Info.Title,
	}

	if mov.List == rms_library.List_WatchList {
		selopts.Criteria = selector.CriteriaFastest
	}

	result, err := strategy.Search(ctx, mov.ID.String(), &mov.Info, selopts)
	if err != nil {
		return fmt.Errorf("search content failed: %w", err)
	}

	if len(result) == 0 {
		return errors.New("nothing found")
	}

	for _, r := range result {
		if err = l.dm.Download(ctx, &mov.ListItem, r.Torrent); err != nil {
			log.Logf(logger.ErrorLevel, "Download failed: %s", err)
		}
	}

	l.notifyUser(log, ctx, mov, events.Notification_ContentFound, getSeasons(result))
	return nil
}

func (l MoviesService) searchAndSave(log logger.Logger, ctx context.Context, mov *model.Movie) error {
	sel := l.getMovieSelector(mov)
	opts := selector.Options{
		Criteria:  selector.CriteriaQuality,
		MediaType: media.Movies,
		Query:     mov.Info.Title,
	}

	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)

	result, err := searchEngine.SearchTorrents(ctx, mov.ID.String(), &mov.Info, nil)
	if err != nil {
		logger.Errorf("Find torrents failed: %s", err)
		return err
	}
	if len(result) == 0 {
		return errors.New("nothing found")
	}

	sel.Sort(result, opts)
	result = boundResults(result)

	mov.ArchivedTorrents = l.fetchTorrentFiles(context.Background(), searchEngine, mov.Info.Title, result)

	if mov.Info.Type == rms_library.MovieType_TvSeries && mov.Info.Seasons != nil {
		opts.Criteria = selector.CriteriaQuality
		mov.ArchivedSeasons = map[uint][]model.TorrentSearchResult{}
		for season := uint(1); season <= uint(*mov.Info.Seasons); season++ {
			result, err = searchEngine.SearchTorrents(context.Background(), mov.ID.String(), &mov.Info, &season)
			if err != nil {
				logger.Errorf("Find torrents failed: %s", err)
				continue
			}
			sel.Sort(result, opts)
			result = boundResults(result)
			mov.ArchivedSeasons[uint(season)] = l.fetchTorrentFiles(context.Background(), searchEngine, mov.Info.Title, result)
			logger.Infof("For %s [ %s ] found season no%.d, torrents: %d", mov.Info.Title, mov.ID, season, len(result))
		}
	}

	if err := l.db.UpdateMovieArchiveContent(context.Background(), mov); err != nil {
		logger.Errorf("Save items of '%s' [ %s ] to archive failed: %s", mov.Info.Title, mov.ID, err)
		return err
	}

	totalSeasons := []uint32{}
	for no := range mov.ArchivedSeasons {
		totalSeasons = append(totalSeasons, uint32(no))
	}
	sort.SliceStable(totalSeasons, func(i, j int) bool { return totalSeasons[i] < totalSeasons[j] })
	l.notifyUser(log, ctx, mov, events.Notification_ContentFound, totalSeasons)

	logger.Infof("Item '%s' [ %s ] saved to archive", mov.Info.Title, mov.ID)
	return nil
}

func (l MoviesService) notifyUser(log logger.Logger, ctx context.Context, mov *model.Movie, kind events.Notification_Kind, seasons []uint32) {
	nCtx, nCancel := context.WithTimeout(ctx, notifyTimeout)
	defer nCancel()

	size := uint32(mov.Size())
	event := events.Notification{
		Sender:    "rms-library",
		Kind:      kind,
		MediaID:   (*string)(&mov.ID),
		ItemTitle: &mov.Title,
		SizeMB:    &size,
		Seasons:   seasons,
	}

	if err := l.pub.Publish(nCtx, &event); err != nil {
		log.Logf(logger.WarnLevel, "Send notification about movie failed: %s", err)
	}
}

func getSeasons(result []movsearch.Result) []uint32 {
	foundSeasons := movsearch.GetMultipleResultsSeasons(result)
	seasons := make([]uint32, 0, len(foundSeasons))
	for s := range seasons {
		seasons = append(seasons, uint32(s))
	}
	sort.SliceStable(seasons, func(i, j int) bool { return seasons[i] < seasons[j] })
	return seasons
}
