package movies

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/internal/schedule"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

// Database requires some methods for load and store data
type Database interface {
	// cache
	PutMovieInfo(ctx context.Context, id model.ID, mov *rms_library.MovieInfo) error
	GetMovieInfo(ctx context.Context, id model.ID) (*rms_library.MovieInfo, error)

	// persistent
	AddMovie(ctx context.Context, mov *model.Movie) error
	GetMovie(ctx context.Context, id model.ID) (*model.Movie, error)
}

type DirectoryManager interface {
	GetDownloadedSeasons(mov *model.Movie) map[uint]struct{}
	StoreWatchListTorrent(itemTitle string, torrent []byte) (id model.ID, err error)
	LoadWatchListTorrent(contentPath string) ([]byte, error)
}

type DownloadsManager interface {
	Download(ctx context.Context, item *model.ListItem, category string, torrent []byte) error
}

type Scheduler interface {
	Add(t *schedule.Task) bool
	Cencel(groupId string)
}
