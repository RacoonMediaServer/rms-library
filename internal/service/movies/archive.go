package movies

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"go-micro.dev/v4/logger"
)

const maxTorrentsInWatchListItem = 5

func boundResults(results []*models.SearchTorrentsResult) []*models.SearchTorrentsResult {
	if len(results) > maxTorrentsInWatchListItem {
		return results[:maxTorrentsInWatchListItem]
	}
	return results
}

func (l MoviesService) fetchTorrentFiles(ctx context.Context, searcher movsearch.SearchEngine, title string, results []*models.SearchTorrentsResult) []model.TorrentSearchResult {
	items := make([]model.TorrentSearchResult, 0, len(results))
	for _, r := range results {
		content, err := searcher.GetTorrentFile(ctx, *r.Link)
		if err != nil {
			logger.Warnf("Download torrent failed: %s", err)
			continue
		}
		pathToTorrentFile, err := l.dir.StoreArchiveTorrent(title, content)
		if err != nil {
			logger.Warnf("Save to watchlist failed: %s", err)
			continue
		}
		item := model.TorrentSearchResult{
			SearchTorrentsResult: *r,
			Path:                 pathToTorrentFile,
		}
		items = append(items, item)
	}

	return items
}
