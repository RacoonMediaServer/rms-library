package downloads

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type Database interface {
	UpdateContent(ctx context.Context, id model.ID, torrents []model.TorrentRecord) error
	SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error)
	GetMovie(ctx context.Context, id model.ID) (*model.Movie, error)
}

type DirectoryManager interface {
	MoviesMountTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord) error
	MoviesUmountTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord)
}
