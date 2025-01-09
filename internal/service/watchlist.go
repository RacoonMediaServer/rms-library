package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/selector"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

const maxTorrentsInWatchListItem = 5

func boundResults(results []*models.SearchTorrentsResult) []*models.SearchTorrentsResult {
	if len(results) > maxTorrentsInWatchListItem {
		return results[:maxTorrentsInWatchListItem]
	}
	return results
}

func (l LibraryService) fetchTorrentFiles(ctx context.Context, title string, results []*models.SearchTorrentsResult) []model.TorrentItem {
	items := make([]model.TorrentItem, 0, len(results))
	for _, r := range results {
		content, err := l.downloadTorrent(ctx, *r.Link)
		if err != nil {
			logger.Warnf("Download torrent failed: %s", err)
			continue
		}
		contentID, err := l.dir.StoreWatchListTorrent(title, content)
		if err != nil {
			logger.Warnf("Save to watchlist failed: %s", err)
			continue
		}
		item := model.TorrentItem{
			SearchTorrentsResult: *r,
			TorrentContent:       contentID,
		}
		items = append(items, item)
	}

	return items
}

func (l LibraryService) WatchLater(ctx context.Context, request *rms_library.WatchLaterRequest, empty *emptypb.Empty) error {
	logger.Infof("WatchLater: %s", request.Id)
	mov, err := l.getOrCreateMovie(ctx, request.Id)
	if err != nil {
		err = fmt.Errorf("get or create movie failed: %s", err)
		logger.Error(err)
		return err
	}

	item := model.WatchListItem{
		ID:        request.Id,
		Type:      media.Movies,
		MovieInfo: mov.Info,
	}

	sel := l.getMovieSelector(mov)
	opts := selector.Options{
		Criteria:  selector.CriteriaQuality,
		MediaType: media.Movies,
		Query:     mov.Info.Title,
	}
	if mov.Info.Type == rms_library.MovieType_TvSeries {
		opts.Criteria = selector.CriteriaCompact
	}

	result, err := l.searchMovieTorrents(ctx, &mov.Info, nil, searchTorrentsLimit)
	if err != nil {
		logger.Errorf("Find torrents failed: %s", err)
		return err
	}
	if len(result) == 0 {
		return errors.New("nothing found")
	}
	sel.Sort(result, opts)
	result = boundResults(result)

	go func() {
		item.Torrents = l.fetchTorrentFiles(context.Background(), mov.Info.Title, result)

		if mov.Info.Type == rms_library.MovieType_TvSeries && mov.Info.Seasons != nil {
			opts.Criteria = selector.CriteriaQuality
			item.Seasons = map[uint][]model.TorrentItem{}
			for season := uint32(1); season <= *mov.Info.Seasons; season++ {
				result, err = l.searchMovieTorrents(context.Background(), &mov.Info, &season, searchTorrentsLimit)
				if err != nil {
					logger.Errorf("Find torrents failed: %s", err)
					continue
				}
				sel.Sort(result, opts)
				result = boundResults(result)
				item.Seasons[uint(season)] = l.fetchTorrentFiles(context.Background(), mov.Info.Title, result)
				logger.Infof("For %s [ %s ] found season no%.d, torrents: %d", mov.Info.Title, mov.ID, season, len(result))
			}
		}

		if err := l.db.AddToWatchList(context.Background(), &item); err != nil {
			logger.Errorf("Save item '%s' [ %s ] to watchlist failed: %s", mov.Info.Title, mov.ID, err)
			return
		}

		logger.Infof("Item '%s' [ %s ] saved to watchlist", mov.Info.Title, mov.ID)
	}()

	return nil
}
