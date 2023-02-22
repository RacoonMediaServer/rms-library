package storage

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"os"
	"path"
)

const mediaPerms = 0755

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
	return nil
}

func (m Manager) createFilmLinks(dir, torrent string, files []model.File) error {
	for _, f := range files {
		if f.Type != model.FileTypeInsignificant {
			_, fileName := path.Split(f.Path)
			oldName := path.Join(m.TorrentsDirectory(), torrent, f.Path)
			newName := path.Join(dir, fileName)
			if err := os.Symlink(oldName, newName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m Manager) createSeasonLinks(dir string, s *model.Season) error {
	for _, e := range s.Episodes {
		if e.Type == model.FileTypeInsignificant {
			continue
		}
		_, fileName := path.Split(e.Path)
		ext := path.Ext(e.Path)

		if e.No > -1 {
			fileName = fmt.Sprintf("S%02dE%02d. %s%s", s.No, e.No, e.Title, ext)
		}
		oldName := path.Join(m.TorrentsDirectory(), s.TorrentID, e.Path)
		newName := path.Join(dir, fileName)
		if err := os.Symlink(oldName, newName); err != nil {
			return err
		}
	}

	return nil
}

// CreateMovieLayout creates pretty symbolic links to movie
func (m Manager) CreateMovieLayout(mov *model.Movie) error {
	dir := path.Join(m.MoviesDirectory(), mov.Info.Title) //
	_ = os.RemoveAll(dir)

	if err := os.MkdirAll(dir, mediaPerms); err != nil {
		return err
	}
	if mov.Info.Type == rms_library.MovieType_TvSeries {
		variants := map[uint]int{}
		for _, season := range mov.Seasons {
			dir := path.Join(dir, fmt.Sprintf("Сезон %d", season.No))
			if mov.HasSeasonVariants(season.No) {
				variants[season.No]++
				dir = path.Join(dir, fmt.Sprintf("Вариант %d", variants[season.No]))
			}
			if err := os.MkdirAll(dir, mediaPerms); err != nil {
				return err
			}
			if err := m.createSeasonLinks(dir, &season); err != nil {
				return err
			}
		}

	} else {
		if len(mov.Files) == 1 {
			torrent, files := getFirst(mov.Files)
			return m.createFilmLinks(dir, torrent, files)
		}
		invariant := 1
		for torrent, files := range mov.Files {
			dir := path.Join(dir, fmt.Sprintf("Вариант %d", invariant))
			invariant++
			if err := os.MkdirAll(dir, mediaPerms); err != nil {
				return err
			}
			if err := m.createFilmLinks(dir, torrent, files); err != nil {
				return err
			}
		}
	}

	return nil
}
