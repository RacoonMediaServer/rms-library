package storage

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"os"
	"path"
)

const mediaPerms = 0755
const downloadPerms = 0777

// Manager is responsible for management content on a disk
type Manager struct {
	// BaseDirectory is an absolute path to media directory on the device
	BaseDirectory string
}

// CreateDefaultLayout creates default folders on media directory
func (m Manager) CreateDefaultLayout() error {
	if err := os.MkdirAll(m.TorrentsDirectory(), 0777); err != nil {
		return fmt.Errorf("create torrents directory failed: %w", err)
	}
	if err := os.MkdirAll(m.MoviesDirectory(), mediaPerms); err != nil {
		return fmt.Errorf("create torrents directory failed: %w", err)
	}
	if err := os.MkdirAll(m.DownloadsDirectory(), downloadPerms); err != nil {
		return fmt.Errorf("create downloads directory failed: %w", err)
	}
	return nil
}

func (m Manager) createFilmLinks(mov *model.Movie, dir string) error {
	for _, f := range mov.Files {
		if f.Type != model.FileTypeInsignificant {
			oldName := path.Join(m.TorrentsDirectory(), mov.TorrentID, f.Path)
			newName := path.Join(dir, composeFileName(mov, &f))
			if err := os.Symlink(oldName, newName); err != nil {
				logger.Warnf("Create link failed: %s", err)
			}
		}
	}

	return nil
}

func (m Manager) createSeasonLinks(mov *model.Movie, dir string, no uint, s *model.Season) error {
	for _, e := range s.Episodes {
		if e.Type == model.FileTypeInsignificant {
			continue
		}
		oldName := path.Join(m.TorrentsDirectory(), s.TorrentID, e.Path)
		newName := path.Join(dir, composeFileName(mov, &e))
		if _, err := os.Stat(oldName); err != nil {
			continue
		}
		if err := os.Symlink(oldName, newName); err != nil {
			logger.Warnf("Create link failed: %s", err)
		}
	}

	return nil
}

// CreateMovieLayout creates pretty symbolic links to movie
func (m Manager) CreateMovieLayout(mov *model.Movie) error {
	dir := path.Join(m.MoviesDirectory(), getCategory(mov), mov.Info.Title)
	_ = os.RemoveAll(dir)

	if err := os.MkdirAll(dir, mediaPerms); err != nil {
		return err
	}
	if mov.Info.Type == rms_library.MovieType_TvSeries {
		for no, season := range mov.Seasons {
			dir := path.Join(dir, fmt.Sprintf("Сезон %d", no))
			if err := os.MkdirAll(dir, mediaPerms); err != nil {
				return err
			}
			if err := m.createSeasonLinks(mov, dir, no, season); err != nil {
				return err
			}
		}

	} else {
		return m.createFilmLinks(mov, dir)
	}

	return nil
}

// DeleteMovieLayout removes all links to the movie
func (m Manager) DeleteMovieLayout(mov *model.Movie) error {
	dir := path.Join(m.MoviesDirectory(), getCategory(mov), mov.Info.Title)
	return os.RemoveAll(dir)
}

// GetMovieFilePath returns relative tv-series or movie file path
func (m Manager) GetMovieFilePath(mov *model.Movie, season uint, f *model.File) string {
	if mov.Info.Type == rms_library.MovieType_Film {
		return path.Join(getCategory(mov), mov.Info.Title, composeFileName(mov, f))
	}
	return path.Join(getCategory(mov), mov.Info.Title, fmt.Sprintf("Сезон %d", season), composeFileName(mov, f))
}

func (m Manager) CreateMoviesLayout(movies []*model.Movie) error {
	dirs, err := os.ReadDir(m.MoviesDirectory())
	if err != nil {
		return err
	}
	for _, d := range dirs {
		_ = os.RemoveAll(path.Join(m.MoviesDirectory(), d.Name()))
	}

	for _, mov := range movies {
		_ = m.CreateMovieLayout(mov)
	}

	return nil
}
