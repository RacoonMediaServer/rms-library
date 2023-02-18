package service

import (
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"sync"
)

type cache struct {
	mu        sync.RWMutex
	movieInfo map[string]*rms_library.MovieInfo
}

func newCache() *cache {
	return &cache{
		movieInfo: make(map[string]*rms_library.MovieInfo),
	}
}

func (c *cache) PutMovieInfo(id string, info *rms_library.MovieInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.movieInfo[id] = info
}

func (c *cache) GetMovieInfo(id string) (*rms_library.MovieInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.movieInfo[id]
	return val, ok
}
