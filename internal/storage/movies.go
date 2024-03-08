package storage

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"os"
	"path"
	"unicode"
	"unicode/utf8"
)

const fixedTorrentDir = "data"

func getMovieDirectories(mov *model.Movie) (directories []string) {
	title := escape(mov.Info.Title)

	directories = append(directories, path.Join(getMovieCategoryDir(mov), title))

	if mov.Info.Year != 0 {
		directories = append(directories, path.Join(nameByYear, fmt.Sprintf("%d", mov.Info.Year), title))
	}

	letter, _ := utf8.DecodeRuneInString(title)
	letter = unicode.ToUpper(letter)
	directories = append(directories, path.Join(nameByAlpha, string(letter), title))

	for _, g := range mov.Info.Genres {
		g = capitalize(g)
		directories = append(directories, path.Join(nameByGenre, g, title))
	}

	return
}

func (m *Manager) getFullFilePath(torrentID, shortPath string) string {
	if m.fixTorrentPath {
		return path.Join(m.TorrentsDirectory(), fixedTorrentDir, shortPath)
	} else {
		return path.Join(m.TorrentsDirectory(), torrentID, shortPath)
	}
}

func (m *Manager) createFilmLinks(mov *model.Movie, dir string) {
	for _, f := range mov.Files {
		if f.Type != model.FileTypeInsignificant {
			oldName := m.getFullFilePath(mov.TorrentID, f.Path)
			newName := path.Join(dir, composeMovieFileName(mov, &f))
			if err := os.Symlink(oldName, newName); err != nil {
				logger.Warnf("Create link failed: %s", err)
			}
		}
	}
}

func (m *Manager) createClipLinks(mov *model.Movie, dir string) {
	for _, f := range mov.Files {
		oldName := m.getFullFilePath(mov.TorrentID, f.Path)
		newName := path.Join(dir, f.Path)
		_ = os.MkdirAll(path.Dir(newName), mediaPerms)
		if err := os.Symlink(oldName, newName); err != nil {
			logger.Warnf("Create link failed: %s", err)
		}
	}
}

func (m *Manager) createSeasonLinks(mov *model.Movie, dir string, no uint, s *model.Season) {
	for _, e := range s.Episodes {
		if e.Type == model.FileTypeInsignificant {
			continue
		}
		oldName := m.getFullFilePath(s.TorrentID, e.Path)
		newName := path.Join(dir, composeMovieFileName(mov, &e))
		if _, err := os.Stat(oldName); err != nil {
			continue
		}
		if err := os.Symlink(oldName, newName); err != nil {
			logger.Warnf("Create link failed: %s", err)
		}
	}
}

// GetMovieFilePath returns relative tv-series or movie file path
func (m *Manager) GetMovieFilePath(mov *model.Movie, season uint, f *model.File) string {
	switch mov.Info.Type {
	case rms_library.MovieType_Film:
		return path.Join(getMovieCategoryDir(mov), mov.Info.Title, composeMovieFileName(mov, f))
	case rms_library.MovieType_TvSeries:
		return path.Join(getMovieCategoryDir(mov), mov.Info.Title, fmt.Sprintf("Сезон %d", season), composeMovieFileName(mov, f))
	case rms_library.MovieType_Clip:
		return path.Join(getMovieCategoryDir(mov), mov.Info.Title, composeMovieFileName(mov, f))
	}
	return ""
}

func (m *Manager) CreateMoviesLayout(movies []*model.Movie) error {
	dirs, err := os.ReadDir(m.MoviesDirectory())
	if err != nil {
		return err
	}
	for _, d := range dirs {
		_ = os.RemoveAll(path.Join(m.MoviesDirectory(), d.Name()))
	}

	for _, mov := range movies {
		m.CreateMovieLayout(mov)
	}

	return nil
}

// CreateMovieLayout creates pretty symbolic links to movie
func (m *Manager) CreateMovieLayout(mov *model.Movie) {
	m.cmd <- func() {
		dirs := getMovieDirectories(mov)

		for _, dir := range dirs {
			dir = path.Join(m.MoviesDirectory(), dir)
			_ = os.RemoveAll(dir)

			if err := os.MkdirAll(dir, mediaPerms); err != nil {
				logger.Warnf("Cannot create directory: %s", err)
				continue
			}
			switch mov.Info.Type {
			case rms_library.MovieType_TvSeries:
				for no, season := range mov.Seasons {
					dir := path.Join(dir, fmt.Sprintf("Сезон %d", no))
					if err := os.MkdirAll(dir, mediaPerms); err != nil {
						logger.Warnf("Cannot create directory: %s", err)
					}
					m.createSeasonLinks(mov, dir, no, season)
				}
			case rms_library.MovieType_Film:
				m.createFilmLinks(mov, dir)
			case rms_library.MovieType_Clip:
				m.createClipLinks(mov, dir)
			}
		}
	}
}

// DeleteMovieLayout removes all links to the movie
func (m *Manager) DeleteMovieLayout(mov *model.Movie) {
	m.cmd <- func() {
		dirs := getMovieDirectories(mov)
		for _, dir := range dirs {
			dir = path.Join(m.MoviesDirectory(), dir)
			_ = os.RemoveAll(dir)
		}
	}
}
