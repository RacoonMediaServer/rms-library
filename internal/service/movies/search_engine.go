package movies

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/go-openapi/runtime"
)

type torrentSearchEngine interface {
	SearchTorrents(ctx context.Context, mov *model.Movie, season *uint32) (result []*models.SearchTorrentsResult, err error)
	GetTorrentFile(ctx context.Context, link string) ([]byte, error)
	SetNext(engine torrentSearchEngine)
}

type remoteSearchEngine struct {
	next torrentSearchEngine

	service torrents.ClientService
	auth    runtime.ClientAuthInfoWriter
}

func newRemoteSearchEngine(service torrents.ClientService, auth runtime.ClientAuthInfoWriter) torrentSearchEngine {
	return &remoteSearchEngine{service: service, auth: auth}
}

type watchListSearchEngine struct {
	next torrentSearchEngine

	db  Database
	dir DirectoryManager
}

func newWatchListSearchEngine(db Database, dir DirectoryManager) torrentSearchEngine {
	return &watchListSearchEngine{
		db:  db,
		dir: dir,
	}
}

func (e *remoteSearchEngine) asyncSearch(ctx context.Context, q torrents.SearchTorrentsAsyncBody) ([]*models.SearchTorrentsResult, error) {
	sess, err := e.service.SearchTorrentsAsync(&torrents.SearchTorrentsAsyncParams{SearchParameters: q, Context: ctx}, e.auth)
	if err != nil {
		return nil, err
	}
	defer e.service.SearchTorrentsAsyncCancel(&torrents.SearchTorrentsAsyncCancelParams{ID: sess.Payload.ID, Context: ctx}, e.auth)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(sess.Payload.PollIntervalMs) * time.Millisecond):
		}
		resp, err := e.service.SearchTorrentsAsyncStatus(&torrents.SearchTorrentsAsyncStatusParams{ID: sess.Payload.ID, Context: ctx}, e.auth)
		if err != nil {
			return nil, err
		}
		switch *resp.Payload.Status {
		case "ready":
			return resp.Payload.Results, nil
		case "error":
			return nil, errors.New(resp.Payload.Error)
		default:
			continue
		}
	}
}

func (e *remoteSearchEngine) searchTorrents(ctx context.Context, mov *model.Movie, season *uint32) (result []*models.SearchTorrentsResult, err error) {
	strong := true
	q := torrents.SearchTorrentsAsyncBody{
		Limit:  int64(searchTorrentsLimit),
		Q:      &mov.Info.Title,
		Type:   "movies",
		Strong: &strong,
	}

	if mov.Info.Type == rms_library.MovieType_TvSeries && season != nil {
		s := int64(*season)
		q.Season = s
	}

	if mov.Info.Type == rms_library.MovieType_Film && mov.Info.Year != 0 {
		q.Year = int64(mov.Info.Year)
	}

	result, err = e.asyncSearch(ctx, q)
	if err == nil && len(result) == 0 && mov.Info.Title != mov.Info.OriginalTitle && mov.Info.OriginalTitle != "" {
		q.Q = &mov.Info.OriginalTitle
		result, err = e.asyncSearch(ctx, q)
	}
	return
}

func (e *remoteSearchEngine) SetNext(next torrentSearchEngine) {
	e.next = next
}

func (e *remoteSearchEngine) SearchTorrents(ctx context.Context, mov *model.Movie, season *uint32) (result []*models.SearchTorrentsResult, err error) {
	result, err = e.searchTorrents(ctx, mov, season)
	if err != nil && e.next != nil {
		result, err = e.next.SearchTorrents(ctx, mov, season)
		return
	}
	if err == nil && len(result) == 0 {
		err = errAnyTorrentsNotFound
	}
	return
}

func (e *remoteSearchEngine) GetTorrentFile(ctx context.Context, link string) ([]byte, error) {
	download := &torrents.DownloadTorrentParams{
		Link:    link,
		Context: ctx,
	}
	buf := bytes.NewBuffer([]byte{})

	_, err := e.service.DownloadTorrent(download, e.auth, buf)
	if err != nil {
		if e.next == nil {
			return nil, fmt.Errorf("download torrent file failed: %w", err)
		}
		return e.next.GetTorrentFile(ctx, link)
	}

	return buf.Bytes(), nil
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

func (e *watchListSearchEngine) searchTorrents(ctx context.Context, mov *model.Movie, season *uint32) ([]*models.SearchTorrentsResult, error) {
	item, err := e.db.GetWatchListItem(ctx, mov.ID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("not found")
	}

	if mov.Info.Type == rms_library.MovieType_Film || season == nil {
		return e.convertTorrents(item.Torrents), nil
	}

	return e.convertTorrents(item.Seasons[uint(*season)]), nil
}

func (e *watchListSearchEngine) SetNext(next torrentSearchEngine) {
	e.next = next
}

func (e *watchListSearchEngine) SearchTorrents(ctx context.Context, mov *model.Movie, season *uint32) (result []*models.SearchTorrentsResult, err error) {
	result, err = e.searchTorrents(ctx, mov, season)
	if err != nil && e.next != nil {
		result, err = e.next.SearchTorrents(ctx, mov, season)
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
