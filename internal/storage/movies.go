package storage

import (
	"context"
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

// CreateMovieLayout creates pretty symbolic links to movie
func (m *Manager) CreateMovieLayout(mov *model.Movie) {
	m.cmd <- func() {
		mi := m.addMovieInfoToCache(mov)
		l := newMovieLayout(mi, mov.Torrents, m.dirs.Content)
		l.make()
	}
}

func (m *Manager) DeleteItemLayout(id model.ID) {
	m.cmd <- func() {
		dirs := []string{}

		m.mu.Lock()
		mi := m.cache[id]
		if mi != nil {
			delete(m.cache, id)
			dirs = mi.directories
		}
		m.mu.Unlock()

		for _, dir := range dirs {
			dir = path.Join(m.dirs.Content, dir)
			_ = os.RemoveAll(dir)
		}
	}
}

func (m *Manager) UpdateItemLayout(id model.ID) {
	m.cmd <- func() {
		m.mu.Lock()
		mi := m.cache[id]
		m.mu.Unlock()

		if mi == nil {
			return
		}

		item, err := m.db.GetListItem(context.Background(), id)
		if err != nil || item == nil {
			return
		}

		l := newMovieLayout(mi, item.Torrents, m.dirs.Content)
		l.make()
	}
}
