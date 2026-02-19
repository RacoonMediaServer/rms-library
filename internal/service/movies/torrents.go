package movies

import (
	"context"
	"errors"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/v3/internal/model"
	"github.com/RacoonMediaServer/rms-library/v3/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

func convertTorrent(result *models.SearchTorrentsResult) *rms_library.Torrent {
	return &rms_library.Torrent{
		Id:      *result.Link,
		Title:   *result.Title,
		Size:    uint64(*result.Size),
		Seeders: uint32(*result.Seeders),
	}
}

func (l MoviesService) GetTorrentContent(ctx context.Context, torrentId string) ([]byte, error) {
	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)
	return searchEngine.GetTorrentFile(ctx, torrentId)
}

func (l MoviesService) FindTorrents(ctx context.Context, id model.ID, season *uint) ([]*rms_library.Torrent, error) {
	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load movie failed: %w", err)
	}
	if mov == nil {
		return nil, errors.New("movie not found")
	}

	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)
	resp, err := searchEngine.SearchTorrents(ctx, mov.ID.String(), &mov.Info, season)
	if err != nil {
		return nil, fmt.Errorf("search torrents failed: %s", err)
	}

	result := make([]*rms_library.Torrent, len(resp))
	for i := range resp {
		result[i] = convertTorrent(resp[i])
	}
	return result, nil
}
