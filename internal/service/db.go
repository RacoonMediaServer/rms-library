package service

import (
	"context"
	"github.com/RacoonMediaServer/rms-library/internal/model"
)

// Database requires some methods for load and store data
type Database interface {
	GetDownloadedSeasons(ctx context.Context, id string) ([]uint, error)
	GetOrCreateMovie(ctx context.Context, mov *model.Movie) error
	UpdateMovieContent(mov *model.Movie) error
}
