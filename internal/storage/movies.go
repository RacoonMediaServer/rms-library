package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/RacoonMediaServer/rms-library/internal/analysis"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

const rawFilesDirectory = "_Raw"

type movieLayout struct {
	root         string
	l            logger.Logger
	mi           *rms_library.MovieInfo
	t            *model.TorrentRecord
	mapMovieDirs []string
}

// MoviesMountTorrent implements downloads.DirectoryManager.
func (m *Manager) MoviesMountTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord) error {
	l := logger.DefaultLogger.Fields(map[string]interface{}{
		"title":   mi.Title,
		"tid":     t.ID,
		"torrent": t.Title,
	})

	fi, err := os.Stat(t.Location)
	if err != nil {
		l.Logf(logger.ErrorLevel, "Location '%s' is empty or inaccessible", t.Location)
		return err
	}

	ml := &movieLayout{
		root:         m.dirs.Content,
		l:            l,
		mi:           mi,
		t:            t,
		mapMovieDirs: mapMovieDirectories(mi, t.Title),
	}

	if !fi.IsDir() {
		ml.makeLinks(t.Location, fi.Name())
		return nil
	}

	ml.mount()
	return nil
}

// MoviesUmountTorrent implements downloads.DirectoryManager.
func (m *Manager) MoviesUmountTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord) {
	l := logger.DefaultLogger.Fields(map[string]interface{}{
		"title":   mi.Title,
		"tid":     t.ID,
		"torrent": t.Title,
	})
	ml := &movieLayout{
		root:         m.dirs.Content,
		l:            l,
		mi:           mi,
		t:            t,
		mapMovieDirs: mapMovieDirectories(mi, t.Title),
	}
	ml.umount()
}

func mapMovieDirectories(mi *rms_library.MovieInfo, torrentTitle string) (directories []string) {
	title := escape(mi.Title)
	torrentTitle = escape(torrentTitle)

	directories = append(directories, path.Join(getMovieCategoryDir(mi), title, torrentTitle))

	if mi.Year != 0 {
		directories = append(directories, path.Join(nameByYear, fmt.Sprintf("%d", mi.Year), title, torrentTitle))
	}

	letter, _ := utf8.DecodeRuneInString(title)
	letter = unicode.ToUpper(letter)
	directories = append(directories, path.Join(nameByAlpha, string(letter), title, torrentTitle))

	for _, g := range mi.Genres {
		g = capitalize(g)
		directories = append(directories, path.Join(nameByGenre, g, title, torrentTitle))
	}

	return
}

func (ml *movieLayout) mount() {
	originDir := ml.t.Location

	err := filepath.Walk(originDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relpath := ""
		relpath, err = filepath.Rel(ml.t.Location, path)
		if err != nil {
			return err
		}
		mediaInfo := analysis.Analyze(relpath)

		ml.makeFileLinks(path, relpath, mediaInfo)

		return nil
	})

	if err != nil {
		ml.l.Logf(logger.ErrorLevel, "Iterate directory '%s' failed: %s", originDir, err)
		return
	}
}

func (ml *movieLayout) makeFileLinks(path, relpath string, result analysis.Result) {
	if ml.mi.Type != rms_library.MovieType_TvSeries {
		ml.makeLinks(path, relpath)
		return
	}

	ml.makeLinks(path, filepath.Join(rawFilesDirectory, relpath))
	if result.Episode == 0 || result.Season == 0 {
		return
	}

	seasonDir := fmt.Sprintf("Сезон %d", result.Season)
	fName := composeMovieFileName(ml.mi, path, &result)
	ml.makeLinks(path, filepath.Join(seasonDir, fName))
}

func (ml *movieLayout) makeLink(origin, target string) {
	_ = os.MkdirAll(path.Dir(target), mediaPerms)
	if err := os.Symlink(origin, target); err != nil {
		ml.l.Logf(logger.WarnLevel, "Create link '%s' -> '%s' failed: %s", target, origin, err)
	}
}

func (ml *movieLayout) makeLinks(origin, target string) {
	for _, dir := range ml.mapMovieDirs {
		absTarget := filepath.Join(ml.root, dir, target)
		ml.makeLink(origin, absTarget)
	}
}

func (ml *movieLayout) umount() {
	for _, dir := range ml.mapMovieDirs {
		_ = os.RemoveAll(filepath.Join(ml.root, dir))
		// TODO: remove empty directories
	}
}
