package movies

import (
	"context"
	"errors"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type archiveSearchEngine struct {
	next movsearch.SearchEngine

	db  Database
	dir DirectoryManager
}

func newArchiveSearchEngine(db Database, dir DirectoryManager) movsearch.SearchEngine {
	return &archiveSearchEngine{
		db:  db,
		dir: dir,
	}
}

func (e *archiveSearchEngine) convertTorrents(torrents []model.TorrentSearchResult) []*models.SearchTorrentsResult {
	result := make([]*models.SearchTorrentsResult, len(torrents))
	for i, t := range torrents {
		result[i] = &t.SearchTorrentsResult
		t.Link = new(string)
		*t.Link = t.Path
	}
	return result
}

func (e *archiveSearchEngine) searchTorrents(ctx context.Context, id string, info *rms_library.MovieInfo, season *uint) ([]*models.SearchTorrentsResult, error) {
	mov, err := e.db.GetMovie(ctx, model.ID(id))
	if err != nil {
		return nil, err
	}
	if mov == nil {
		return nil, errors.New("not found")
	}

	if info.Type == rms_library.MovieType_Film || season == nil {
		return e.convertTorrents(mov.ArchivedTorrents), nil
	}

	return e.convertTorrents(mov.ArchivedSeasons[*season]), nil
}

func (e *archiveSearchEngine) SetNext(next movsearch.SearchEngine) {
	e.next = next
}

func (e *archiveSearchEngine) SearchTorrents(ctx context.Context, id string, info *rms_library.MovieInfo, season *uint) (result []*models.SearchTorrentsResult, err error) {
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

func (e *archiveSearchEngine) GetTorrentFile(ctx context.Context, link string) ([]byte, error) {
	data, err := e.dir.LoadWatchListTorrent(link)
	if err != nil && e.next != nil {
		return e.next.GetTorrentFile(ctx, link)
	}
	return data, err
}
