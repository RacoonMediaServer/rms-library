package movies

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

// Database requires some methods for load and store data
type Database interface {
	GetOrCreateMovie(ctx context.Context, mov *model.Movie) error
	GetMovie(ctx context.Context, id string) (*model.Movie, error)
	SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error)
	UpdateAvailableContent(ctx context.Context, mov *model.Movie) error
	DeleteMovie(ctx context.Context, id string) error

	PutMovieInfo(ctx context.Context, id string, mov *rms_library.MovieInfo) error
	GetMovieInfo(ctx context.Context, id string) (*rms_library.MovieInfo, error)

	AddToWatchList(ctx context.Context, item *model.WatchListItem) error
	GetWatchList(ctx context.Context, movieType *rms_library.MovieType) ([]*model.WatchListItem, error)
	GetWatchListItem(ctx context.Context, id string) (*model.WatchListItem, error)
}

type DirectoryManager interface {
	GetDownloadedSeasons(mov *model.Movie) map[uint]struct{}
	StoreWatchListTorrent(itemTitle string, torrent []byte) (id string, err error)
	LoadWatchListTorrent(contentPath string) ([]byte, error)
}

type DownloadsManager interface {
	DownloadMovie(ctx context.Context, mov *model.Movie, voice string, torrent []byte, online bool) error
	RemoveMovie(ctx context.Context, mov *model.Movie) error
	GetMovieByTorrent(torrentID string) (string, bool)
	HandleTorrentEvent(kind events.Notification_Kind, torrentID string, mov *model.Movie)
	GetMovieStoreSize(ctx context.Context, mov *model.Movie) uint64
}
