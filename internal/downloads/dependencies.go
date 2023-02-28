package downloads

import (
	"context"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

// Database is a database dependency
type Database interface {
	SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error)
}

// DirectoryManager dependency for mapping torrents to media directories
type DirectoryManager interface {
	CreateMoviesLayout(movies []*model.Movie) error
	DeleteMovieLayout(mov *model.Movie) error
	CreateMovieLayout(mov *model.Movie) error
}
