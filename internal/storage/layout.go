package storage

import (
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
	mov              *model.Movie
	downloadsRootDir string
	reprRootDir      string
	rawFolders       map[string][]dirEntry
	filmFile         dirEntry
	tvSeries         tvSeriesLayout
}

func newMovieLayout(mov *model.Movie, downloadsDir, reprDir string) *movieLayout {
	return &movieLayout{
		mov:              mov,
		downloadsRootDir: downloadsDir,
		reprRootDir:      reprDir,
		rawFolders:       map[string][]dirEntry{},
		tvSeries: tvSeriesLayout{
			seasons: map[uint]map[int]dirEntry{},
		},
	}
}

func (l *movieLayout) buildIndex() {
	for _, t := range l.mov.Torrents {
		torrentDir := filepath.Join(model.GetCategory(l.mov.Info.Type), t.Title)
		logger.Warnf("Scan directory %s", torrentDir)
		l.addContentDirectory(torrentDir)
	}
}

func (l *movieLayout) addContentDirectory(dir string) {
	fullDir := filepath.Join(l.downloadsRootDir, dir)
	files := []dirEntry{}
	err := filepath.Walk(fullDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relpath := ""
		relpath, err = filepath.Rel(l.downloadsRootDir, path)
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
		if l.mov.Info.Type == rms_library.MovieType_Film {
			if entry.info.FileType == model.FileTypeFilm {
				l.filmFile = entry
				logger.Debugf("Found film entry '%s' in %s", entry.info.EpisodeName, entry.relpath)
			}
		} else if l.mov.Info.Type == rms_library.MovieType_TvSeries {
			l.tvSeries.addFile(entry)
		}
		return nil
	})
	l.rawFolders[dir] = files

	if err != nil {
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
	dirs := getMovieDirectories(l.mov)

	for _, dir := range dirs {
		_ = os.RemoveAll(path.Join(l.reprRootDir, dir))

		if err := os.MkdirAll(dir, mediaPerms); err != nil {
			logger.Warnf("Cannot create directory: %s", err)
			continue
		}

		switch l.mov.Info.Type {
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
	newName := path.Join(l.reprRootDir, dir, composeMovieFileName(l.mov, &entry))
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
	for _, folder := range l.rawFolders {
		for _, entry := range folder {
			l.makeLink(dir, entry)
		}
	}
}

func (l *movieLayout) makeSeasonLinks(dir string, no uint, season map[int]dirEntry) {
	for _, episode := range season {
		l.makeLink(dir, episode)
	}
}
