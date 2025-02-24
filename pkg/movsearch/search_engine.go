package movsearch

import (
	"context"
	"errors"

	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

const SearchTorrentsLimit uint = 10

var ErrAnyTorrentsNotFound = errors.New("any torrents not found")

type SearchEngine interface {
	SearchTorrents(ctx context.Context, id string, info *rms_library.MovieInfo, season *uint) (result []*models.SearchTorrentsResult, err error)
	GetTorrentFile(ctx context.Context, link string) ([]byte, error)
	SetNext(next SearchEngine)
}
