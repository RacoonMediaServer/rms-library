package service

import (
	"context"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

// Database requires some methods for load and store data
type Database interface {
	GetDownloadedSeasons(ctx context.Context, id string) ([]uint, error)
	GetOrCreateMovie(ctx context.Context, mov *model.Movie) error
	GetMovie(ctx context.Context, id string) (*model.Movie, error)
	UpdateMovieContent(mov *model.Movie) error
	SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error)
	DeleteMovie(ctx context.Context, id string) error

	PutMovieInfo(ctx context.Context, id string, mov *rms_library.MovieInfo) error
	GetMovieInfo(ctx context.Context, id string) (*rms_library.MovieInfo, error)
}
