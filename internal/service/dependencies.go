package service

import (
	"context"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

// Database requires some methods for load and store data
type Database interface {
	GetDownloadedSeasons(ctx context.Context, id string) ([]uint, error)
	GetOrCreateMovie(ctx context.Context, mov *model.Movie) error
	GetMovie(ctx context.Context, id string) (*model.Movie, error)
	SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error)
	UpdateAvailableContent(ctx context.Context, mov *model.Movie) error

	PutMovieInfo(ctx context.Context, id string, mov *rms_library.MovieInfo) error
	GetMovieInfo(ctx context.Context, id string) (*rms_library.MovieInfo, error)
}

type DirectoryManager interface {
	GetFilmFilePath(title string, f *model.File) string
	GetTvSeriesFilePath(title string, season uint, f *model.File) string
}

type DownloadsManager interface {
	DownloadMovie(ctx context.Context, mov *model.Movie, torrent []byte, faster bool) error
	RemoveMovie(ctx context.Context, mov *model.Movie) error
	GetMovieByTorrent(torrentID string) (string, bool)
	HandleTorrentEvent(kind events.Notification_Kind, torrentID string, mov *model.Movie)
}
