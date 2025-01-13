package storage

import (
	"path/filepath"

	"github.com/RacoonMediaServer/rms-library/internal/analysis"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type movieLayout struct {
	contentRootDir string
	reprRootDir    string
	rawFolders     map[string][]analysis.Result
	belongsTo      *rms_library.MovieType
}

func newMovieLayout(contentDir, reprDir string) *movieLayout {
	return &movieLayout{
		contentRootDir: contentDir,
		reprRootDir:    reprDir,
		rawFolders:     map[string][]analysis.Result{},
	}
}

func (l *movieLayout) addContentDirectory(dir string) {
	fullDir := filepath.Join(l.contentRootDir, contentDir)
	files := []analysis.Result{}
}
