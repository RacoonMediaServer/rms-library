package storage

import (
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type mediaInfo struct {
	directories []string
	movieType   rms_library.MovieType
}

func (m *Manager) addToCache(id model.ID, mi *mediaInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[id] = mi
}

func (m *Manager) addMovieInfoToCache(mov *model.Movie) *mediaInfo {
	mi := mediaInfo{
		directories: getMovieDirectories(mov),
		movieType:   mov.Info.Type,
	}

	m.addToCache(mov.ID, &mi)
	return &mi
}
