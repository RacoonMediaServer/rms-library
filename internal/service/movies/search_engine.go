package movies

import (
	"context"
	"errors"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type watchListSearchEngine struct {
	next movsearch.SearchEngine

	db  Database
	dir DirectoryManager
}

func newWatchListSearchEngine(db Database, dir DirectoryManager) movsearch.SearchEngine {
	return &watchListSearchEngine{
		db:  db,
		dir: dir,
	}
}

func (e *watchListSearchEngine) convertTorrents(torrents []model.TorrentItem) []*models.SearchTorrentsResult {
	result := make([]*models.SearchTorrentsResult, len(torrents))
	for i, t := range torrents {
		result[i] = &t.SearchTorrentsResult
		t.Link = new(string)
		*t.Link = t.TorrentContent
	}
	return result
}

func (e *watchListSearchEngine) searchTorrents(ctx context.Context, id string, info *rms_library.MovieInfo, season *uint) ([]*models.SearchTorrentsResult, error) {
	item, err := e.db.GetWatchListItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("not found")
	}

	if info.Type == rms_library.MovieType_Film || season == nil {
		return e.convertTorrents(item.Torrents), nil
	}

	return e.convertTorrents(item.Seasons[*season]), nil
}

func (e *watchListSearchEngine) SetNext(next movsearch.SearchEngine) {
	e.next = next
}

func (e *watchListSearchEngine) SearchTorrents(ctx context.Context, id string, info *rms_library.MovieInfo, season *uint) (result []*models.SearchTorrentsResult, err error) {
	result, err = e.searchTorrents(ctx, id, info, season)
	if err != nil && e.next != nil {
		result, err = e.next.SearchTorrents(ctx, id, info, season)
		return
	}
	if err == nil && len(result) == 0 {
		err = errAnyTorrentsNotFound
	}
	return
}

func (e *watchListSearchEngine) GetTorrentFile(ctx context.Context, link string) ([]byte, error) {
	data, err := e.dir.LoadWatchListTorrent(link)
	if err != nil && e.next != nil {
		return e.next.GetTorrentFile(ctx, link)
	}
	return data, err
}
