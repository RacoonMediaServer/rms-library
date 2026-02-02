package downloads

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
)

type Database interface {
	UpdateContent(ctx context.Context, id model.ID, torrents []model.TorrentRecord) error
}

type DirectoryManager interface {
	GetDownloadedSeasons(mov *model.Movie) map[uint]struct{}
	CreateMovieLayout(mov *model.Movie)
	DeleteMovieLayout(mov *model.Movie)
	CreateMoviesLayout(movies []*model.Movie) error
}
