package movsearch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/go-openapi/runtime"
)

type remoteSearchEngine struct {
	next SearchEngine

	service torrents.ClientService
	auth    runtime.ClientAuthInfoWriter
}

func NewRemoteSearchEngine(service torrents.ClientService, auth runtime.ClientAuthInfoWriter) SearchEngine {
	return &remoteSearchEngine{service: service, auth: auth}
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

func (e *remoteSearchEngine) searchTorrents(ctx context.Context, info *rms_library.MovieInfo, season *uint) (result []*models.SearchTorrentsResult, err error) {
	strong := true
	q := torrents.SearchTorrentsAsyncBody{
		Limit:  int64(SearchTorrentsLimit),
		Q:      &info.Title,
		Type:   "movies",
		Strong: &strong,
	}

	if info.Type == rms_library.MovieType_TvSeries && season != nil {
		s := int64(*season)
		q.Season = s
	}

	if info.Type == rms_library.MovieType_Film && info.Year != 0 {
		q.Year = int64(info.Year)
	}

	result, err = e.asyncSearch(ctx, q)
	if err == nil && len(result) == 0 && info.Title != info.OriginalTitle && info.OriginalTitle != "" {
		q.Q = &info.OriginalTitle
		result, err = e.asyncSearch(ctx, q)
	}
	if err == nil && len(result) == 0 {
		strong = false
		result, err = e.asyncSearch(ctx, q)
	}
	return
}

func (e *remoteSearchEngine) SetNext(next SearchEngine) {
	e.next = next
}

func (e *remoteSearchEngine) SearchTorrents(ctx context.Context, id string, info *rms_library.MovieInfo, season *uint) (result []*models.SearchTorrentsResult, err error) {
	result, err = e.searchTorrents(ctx, info, season)
	if err != nil && e.next != nil {
		result, err = e.next.SearchTorrents(ctx, id, info, season)
		return
	}
	if err == nil && len(result) == 0 {
		err = ErrAnyTorrentsNotFound
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
