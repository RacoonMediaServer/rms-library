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
	SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error)
	AddMovie(ctx context.Context, mov *model.Movie) error
	GetMovie(ctx context.Context, id model.ID) (*model.Movie, error)
	UpdateMovieArchiveContent(ctx context.Context, mov *model.Movie) error
	UpdateMovieInfoSeasons(ctx context.Context, mov *model.Movie) error
}

type DirectoryManager interface {
	StoreArchiveTorrent(itemTitle string, torrent []byte) (path string, err error)
	LoadArchiveTorrent(contentPath string) ([]byte, error)
}

type DownloadsManager interface {
	Download(ctx context.Context, item *model.ListItem, torrent []byte) error
	RemoveTorrent(ctx context.Context, item *model.ListItem, torrentId string) error
	DropMissedTorrents(ctx context.Context, item *model.ListItem) error
	UpdateTorrentInfo(ctx context.Context, item *model.ListItem) error
}

type Scheduler interface {
	Add(t *schedule.Task) bool
	Cancel(groupId string)
}
