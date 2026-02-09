package movies

import (
	"context"
	"errors"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/movsearch"
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

func (l MoviesService) FindTorrents(ctx context.Context, id model.ID, torrentId *string) ([]*rms_library.Torrent, error) {
	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load movie failed: %w", err)
	}
	if mov == nil {
		return nil, errors.New("movie not found")
	}

	var t *model.TorrentRecord
	if torrentId != nil {
		t := mov.GetTorrent(*torrentId)
		if t != nil {
			return nil, errors.New("torrent not found")
		}
	}

	if mov.Info.Type != rms_library.MovieType_TvSeries {
		return l.findMovieTorrents(ctx, mov)
	}

	return l.findTvSeriesTorrents(ctx, mov, t)
}

func (l MoviesService) findMovieTorrents(ctx context.Context, mov *model.Movie) ([]*rms_library.Torrent, error) {
	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)

	resp, err := searchEngine.SearchTorrents(ctx, mov.ID.String(), &mov.Info, nil)
	if err != nil {
		return nil, fmt.Errorf("search torrents failed: %s", err)
	}

	result := make([]*rms_library.Torrent, len(resp))
	for i := range resp {
		result[i] = convertTorrent(resp[i])
	}
	return result, nil
}

func (l MoviesService) findTvSeriesTorrents(ctx context.Context, mov *model.Movie, t *model.TorrentRecord) ([]*rms_library.Torrent, error) {
	if t == nil {
		return l.findMovieTorrents(ctx, mov)
	}

	seasons := l.dir.GetTorrentSeasons(t)
	if len(seasons) != 1 {
		return l.findMovieTorrents(ctx, mov)
	}

	var season uint
	for s := range seasons {
		season = s
	}

	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)
	resp, err := searchEngine.SearchTorrents(ctx, mov.ID.String(), &mov.Info, &season)
	if err != nil {
		return nil, fmt.Errorf("search torrents failed: %s", err)
	}

	result := make([]*rms_library.Torrent, len(resp))
	for i := range resp {
		result[i] = convertTorrent(resp[i])
	}
	return result, nil
}
