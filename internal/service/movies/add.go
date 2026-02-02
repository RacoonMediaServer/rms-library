package movies

import (
	"context"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/internal/schedule"
	"github.com/RacoonMediaServer/rms-library/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-library/pkg/selector"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

func (l MoviesService) Add(ctx context.Context, id model.ID, list rms_library.List) error {
	info, err := l.db.GetMovieInfo(ctx, id)
	if err != nil {
		return fmt.Errorf("movie '%s' not found in cache: %w", err)
	}

	mov := model.Movie{
		ListItem: model.ListItem{
			ID:    id,
			Title: info.Title,
			List:  list,
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

func (l MoviesService) asyncDownloadContent(log logger.Logger, ctx context.Context, id model.ID) error {
	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return fmt.Errorf("load movie from database failed: %w", err)
	}

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

	if mov.List != rms_library.List_Archive {
		for _, r := range result {
			if err = l.dm.Download(ctx, &mov.ListItem, model.GetVideoCategory(mov.Info.Type), r.Torrent); err != nil {
				log.Logf(logger.ErrorLevel, "Download failed: %s", err)
			}
		}
	}

	// TODO: send notification

	// seasons := movsearch.GetMultipleResultsSeasons(result)
	// for s := range seasons {
	// 	response.Seasons = append(response.Seasons, uint32(s))
	// }
	// sort.SliceStable(response.Seasons, func(i, j int) bool { return response.Seasons[i] < response.Seasons[j] })
	// response.Found = true

	// TODO: start monitoring task
	// TODO: start grab torrents task
	return nil
}
