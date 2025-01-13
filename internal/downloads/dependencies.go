package downloads

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type Database interface {
	SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error)
	UpdateMovieContent(ctx context.Context, mov *model.Movie) error
	DeleteMovie(ctx context.Context, id string) error
}

type DirectoryManager interface {
	GetDownloadedSeasons(mov *model.Movie) map[uint]struct{}
	CreateMovieLayout(mov *model.Movie)
	DeleteMovieLayout(mov *model.Movie)
	CreateMoviesLayout(movies []*model.Movie) error
}
