package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/RacoonMediaServer/rms-library/internal/analysis"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"go-micro.dev/v4/logger"
)

func (m *Manager) GetDownloadedSeasons(mov *model.Movie) map[uint]struct{} {
	seasons := map[uint]struct{}{}
	for _, t := range mov.Torrents {
		contentPath := t.Location
		err := filepath.Walk(contentPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}

				result := analysis.Analyze(path)
				if result.Season != 0 {
					seasons[result.Season] = struct{}{}
				}
				return nil
			})
		if err != nil {
			logger.Warnf("Walk through %s failed: %s", contentPath, err)
		}
	}
	return seasons
}

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

func (m *Manager) CreateMoviesLayout(movies []*model.Movie) error {
	dirs, err := os.ReadDir(m.dirs.Content)
	if err != nil {
		return err
	}
	for _, d := range dirs {
		_ = os.RemoveAll(path.Join(m.dirs.Content, d.Name()))
	}

	for _, mov := range movies {
		m.CreateMovieLayout(mov)
	}

	return nil
}

// CreateMovieLayout creates pretty symbolic links to movie
func (m *Manager) CreateMovieLayout(mov *model.Movie) {
	m.cmd <- func() {
		l := newMovieLayout(mov, m.dirs.Downloads, m.dirs.Content)
		l.make()
	}
}

// DeleteMovieLayout removes all links to the movie
func (m *Manager) DeleteMovieLayout(mov *model.Movie) {
	m.cmd <- func() {
		dirs := getMovieDirectories(mov)
		for _, dir := range dirs {
			dir = path.Join(m.dirs.Content, dir)
			_ = os.RemoveAll(dir)
		}
	}
}

func (m *Manager) GetTorrentSeasons(t *model.TorrentRecord) map[uint]struct{} {
	seasons := map[uint]struct{}{}
	err := filepath.Walk(t.Location,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			result := analysis.Analyze(path)
			if result.Season != 0 {
				seasons[result.Season] = struct{}{}
			}
			return nil
		})
	if err != nil {
		logger.Warnf("Walk through %s failed: %s", t.Location, err)
	}
	return seasons
}
