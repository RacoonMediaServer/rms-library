package movies

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/lock"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/internal/schedule"
	"github.com/RacoonMediaServer/rms-library/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-library/pkg/selector"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/movies"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

const lockWait = 5 * time.Second

func isContentMissing(mov *model.Movie) bool {
	switch mov.List {
	case rms_library.List_Archive:
		if mov.Info.Type == rms_library.MovieType_Film {
			return len(mov.ArchivedTorrents) == 0
		} else if mov.Info.Type == rms_library.MovieType_TvSeries {
			return len(mov.ArchivedSeasons) == 0
		} else {
			return false
		}
	case rms_library.List_Favourites:
		for _, t := range mov.Torrents {
			if !t.Online {
				return false
			}
		}
		return true
	case rms_library.List_WatchList:
		for _, t := range mov.Torrents {
			if t.Online {
				return false
			}
		}
		return true
	}
	return false
}

func (l MoviesService) startWatchers(mov *model.Movie) {
	// periodic task for validate movie record
	task := schedule.Task{
		Group: mov.ID.String(),
		Fn: schedule.GetPeriodicWrapper(
			logger.Fields(map[string]interface{}{
				"op":    "movieWatcher",
				"id":    mov.ID.String(),
				"title": mov.Info.Title,
			}),
			watchInterval,
			func(log logger.Logger, ctx context.Context) error {
				return l.asyncWatch(log, ctx, mov.ID)
			},
		),
	}
	task.After(time.Duration(rand.Intn(10)) * time.Second)
	l.sched.Add(&task)

	if mov.Info.Type == rms_library.MovieType_TvSeries {
		// periodic task for search new releases
		schedTask := schedule.Task{
			Group: mov.ID.String(),
			Fn: schedule.GetPeriodicWrapper(
				logger.Fields(map[string]interface{}{
					"op":    "movieCheckReleasesWatcher",
					"id":    mov.ID.String(),
					"title": mov.Info.Title,
				}),
				checkReleasesInterval,
				func(log logger.Logger, ctx context.Context) error {
					return l.asyncCheckReleases(log, ctx, mov.ID)
				},
			),
		}
		schedTask.After(time.Duration(rand.Intn(24)) * time.Hour)
		l.sched.Add(&schedTask)
	}
}

func (l MoviesService) asyncWatch(log logger.Logger, ctx context.Context, id model.ID) error {
	lk, err := lock.TimedLock(ctx, l.lk, id, lockWait)
	if err != nil {
		return fmt.Errorf("Lock item failed: %w", err)
	}
	defer lk.Unlock()

	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return fmt.Errorf("load movie from database failed: %w", err)
	}
	if mov == nil {
		return errors.New("movie not found")
	}

	// проверяем различные проблемы и рассинхрон
	// каждую найденную ситуацию разруливаем

	// 1) если какие то торренты пропали с диска - корректируем внутреннее хранилище
	if mov.List != rms_library.List_Archive {
		l.watcherRemoveMissedTorrents(log, ctx, mov)
	}

	// 2) сихронизируем контент и тип списка, в который добавлен item - удаляем торренты, которые считаем лишними
	l.watcherRemoveUnusedTorrents(log, ctx, mov)

	// 3) запускаем загрузку если полностью отсутствует контент
	if isContentMissing(mov) {
		log.Logf(logger.WarnLevel, "Content is missing, try to download all")
		return l.downloadContent(log, ctx, mov)
	}

	// 4) синхронизируем информацию о торрентах
	l.watcherSyncTorrentInfo(log, ctx, mov)

	return nil
}

func (l MoviesService) watcherRemoveUnusedTorrents(log logger.Logger, ctx context.Context, mov *model.Movie) bool {
	changed := false
	removeUnusedTorrent := func(log logger.Logger, ctx context.Context, mov *model.Movie, t *model.TorrentRecord) {
		log.Logf(logger.DebugLevel, "Unused torrent found: %s [ %s ]", t.Title, t.ID)
		if err := l.dm.RemoveTorrent(ctx, &mov.ListItem, t.ID); err != nil {
			log.Logf(logger.WarnLevel, "Remove unused torrent failed: %s", err)
		} else {
			changed = true
		}
	}
	switch mov.List {
	case rms_library.List_Archive:
		for _, t := range mov.Torrents {
			removeUnusedTorrent(log, ctx, mov, &t)
		}
	case rms_library.List_WatchList:
		for _, t := range mov.Torrents {
			if !t.Online {
				removeUnusedTorrent(log, ctx, mov, &t)
			}
		}
	case rms_library.List_Favourites:
		for _, t := range mov.Torrents {
			if t.Online {
				removeUnusedTorrent(log, ctx, mov, &t)
			}
		}
	}

	return changed
}

func (l MoviesService) watcherRemoveMissedTorrents(log logger.Logger, ctx context.Context, mov *model.Movie) {
	if err := l.dm.DropMissedTorrents(ctx, &mov.ListItem); err != nil {
		log.Logf(logger.WarnLevel, "Drop missed torrents failed: %s", err)
	}
}

func (l MoviesService) watcherSyncTorrentInfo(log logger.Logger, ctx context.Context, mov *model.Movie) {
	if err := l.dm.UpdateTorrentInfo(ctx, &mov.ListItem); err != nil {
		log.Logf(logger.WarnLevel, "Sync torrents infromation failed: %s", err)
	}
}

func (l MoviesService) asyncCheckReleases(log logger.Logger, ctx context.Context, id model.ID) error {
	lk, err := lock.TimedLock(ctx, l.lk, id, lockWait)
	if err != nil {
		return fmt.Errorf("Lock item failed: %w", err)
	}
	defer lk.Unlock()

	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return fmt.Errorf("load movie from database failed: %w", err)
	}
	if mov == nil {
		return errors.New("movie not found")
	}

	info, err := l.cli.Movies.GetMovieInfo(&movies.GetMovieInfoParams{ID: mov.ID.Strip(), Context: context.Background()}, l.auth)
	if err != nil {
		return err
	}
	if info.Payload.Seasons == 0 || mov.Info.Seasons == nil {
		return nil
	}
	if *mov.Info.Seasons >= uint32(info.Payload.Seasons) {
		return nil
	}

	newSeasonsCount := uint32(info.Payload.Seasons) - *mov.Info.Seasons
	log.Logf(logger.InfoLevel, "Found info about %d new seasons", newSeasonsCount)

	sel := l.getMovieSelector(mov)
	opts := selector.Options{
		Criteria:  selector.CriteriaQuality,
		MediaType: media.Movies,
		Query:     mov.Info.Title,
	}
	if mov.List == rms_library.List_WatchList {
		opts.Criteria = selector.CriteriaFastest
	}

	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)

	foundRealeses := []uint32{}
	for no := uint(*mov.Info.Seasons); no < uint(info.Payload.Seasons); no++ {
		result, err := searchEngine.SearchTorrents(ctx, mov.ID.String(), &mov.Info, &no)
		if err != nil {
			log.Logf(logger.WarnLevel, "Find torrents for %d season failed: %s", no, err)
			break
		}
		if len(result) == 0 {
			break
		}

		log.Logf(logger.InfoLevel, "Found new releases of %d season!", no)

		if mov.List != rms_library.List_Archive {
			selected := sel.Select(result, opts)
			torrentFile, err := searchEngine.GetTorrentFile(ctx, *selected.Link)
			if err != nil {
				log.Logf(logger.WarnLevel, "Get torrent file of %d season failed: %s", no, err)
				break
			}
			if err := l.dm.Download(ctx, &mov.ListItem, torrentFile); err != nil {
				log.Logf(logger.WarnLevel, "Download new season failed: %s", err)
				break
			}
		} else {
			sel.Sort(result, opts)
			result = boundResults(result)
			mov.ArchivedSeasons[uint(no)] = l.fetchTorrentFiles(ctx, searchEngine, mov.Info.Title, result)
			if err := l.db.UpdateMovieArchiveContent(ctx, mov); err != nil {
				log.Logf(logger.WarnLevel, "update archive failed: %s", err)
				break
			}
		}
		foundRealeses = append(foundRealeses, uint32(no))
	}

	if len(foundRealeses) == 0 {
		return nil
	}

	*mov.Info.Seasons += uint32(len(foundRealeses))
	if err = l.db.UpdateMovieInfoSeasons(ctx, mov); err != nil {
		return fmt.Errorf("update seasons count in database failed: %s", err)
	}

	log.Logf(logger.InfoLevel, "New releases %+v added!", foundRealeses)
	l.notifyUser(log, ctx, mov, events.Notification_NewContentReleased, foundRealeses)
	return nil
}
