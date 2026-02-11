package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/RacoonMediaServer/rms-library/internal/analysis"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

const rawFilesDirectory = "_Torrents"

type dirEntry struct {
	path    string
	relpath string
	info    analysis.Result
}

type tvSeriesLayout struct {
	seasons map[uint]map[int]dirEntry
}

type movieLayout struct {
	mi          *mediaInfo
	reprRootDir string
	rawFolders  map[string][]dirEntry
	filmFile    dirEntry
	tvSeries    tvSeriesLayout
	torrents    []model.TorrentRecord
}

func newMovieLayout(mi *mediaInfo, torrents []model.TorrentRecord, reprDir string) *movieLayout {
	return &movieLayout{
		mi:          mi,
		torrents:    torrents,
		reprRootDir: reprDir,
		rawFolders:  map[string][]dirEntry{},
		tvSeries: tvSeriesLayout{
			seasons: map[uint]map[int]dirEntry{},
		},
	}
}

func (l *movieLayout) buildIndex() {
	for _, t := range l.torrents {
		l.addContentDirectory(t.Title, &t)
	}
}

func (l *movieLayout) addContentDirectory(title string, t *model.TorrentRecord) {
	fullDir := t.Location

	fi, err := os.Stat(t.Location)
	if err != nil {
		logger.Warnf("Stat %s failed: %s", t.Location, err)
		return
	}

	if !fi.IsDir() {
		fullDir = filepath.Dir(t.Location)
	}

	files := []dirEntry{}
	err = filepath.Walk(fullDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relpath := ""
		relpath, err = filepath.Rel(filepath.Dir(t.Location), path)
		if err != nil {
			return err
		}

		entry := dirEntry{
			path:    path,
			relpath: relpath,
			info:    analysis.Analyze(relpath),
		}
		logger.Debugf("File '%s' as '%s' to index", path, relpath)

		files = append(files, entry)
		if l.mi.movieType == rms_library.MovieType_Film {
			if entry.info.FileType == model.FileTypeFilm {
				l.filmFile = entry
				logger.Debugf("Found film entry '%s' in %s", entry.info.EpisodeName, entry.relpath)
			}
		} else if l.mi.movieType == rms_library.MovieType_TvSeries {
			l.tvSeries.addFile(entry)
		}
		return nil
	})
	l.rawFolders[title] = files

	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		logger.Errorf("Iterate directory '%s' failed: %s", fullDir, err)
		return
	}
}

func (l *tvSeriesLayout) addFile(f dirEntry) {
	if f.info.Season == 0 || f.info.Episode < 0 {
		return
	}
	season := l.seasons[f.info.Season]
	if season == nil {
		season = map[int]dirEntry{}
	}
	season[f.info.Episode] = f
	l.seasons[f.info.Season] = season
	logger.Debugf("Found episode '%d.%s' in %s", f.info.Episode, f.info.EpisodeName, f.relpath)
}

func (l *movieLayout) make() {
	l.buildIndex()

	for _, dir := range l.mi.directories {
		_ = os.RemoveAll(path.Join(l.reprRootDir, dir))

		if err := os.MkdirAll(path.Join(l.reprRootDir, dir), mediaPerms); err != nil {
			logger.Warnf("Cannot create directory: %s", err)
			continue
		}

		switch l.mi.movieType {
		case rms_library.MovieType_TvSeries:
			for no, season := range l.tvSeries.seasons {
				l.makeSeasonLinks(dir, no, season)
			}

			l.makeRawLinks(filepath.Join(dir, rawFilesDirectory))
		case rms_library.MovieType_Film:
			l.makeFilmLink(dir)
			l.makeRawLinks(filepath.Join(dir, rawFilesDirectory))
		case rms_library.MovieType_Clip:
			l.makeRawLinks(dir)
		}
	}
}

func (l *movieLayout) makeLink(dir string, entry dirEntry) {
	oldName := entry.path
	newName := path.Join(l.reprRootDir, dir, composeMovieFileName(l.mi, &entry))
	_ = os.MkdirAll(path.Dir(newName), mediaPerms)
	if err := os.Symlink(oldName, newName); err != nil {
		logger.Warnf("Create link failed: %s", err)
	}
}

func (l *movieLayout) makeFilmLink(dir string) {
	if l.filmFile.path == "" {
		return
	}
	l.makeLink(dir, l.filmFile)
}

func (l *movieLayout) makeRawLinks(dir string) {
	for _, files := range l.rawFolders {
		for _, entry := range files {
			newName := path.Join(l.reprRootDir, dir, entry.relpath)
			_ = os.MkdirAll(path.Dir(newName), mediaPerms)
			if err := os.Symlink(entry.path, newName); err != nil {
				logger.Warnf("Create raw link failed: %s", err)
			}
		}
	}
}

func (l *movieLayout) makeSeasonLinks(dir string, no uint, season map[int]dirEntry) {
	for _, episode := range season {
		l.makeLink(filepath.Join(dir, fmt.Sprintf("Сезон %d", no)), episode)
	}
}
