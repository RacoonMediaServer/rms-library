package storage

import (
	"path"
)

const torrentsDirectory = "torrents"

const moviesDirectory = "movies"

// TorrentsDirectory returns absolute path to torrents directory
func (m Manager) TorrentsDirectory() string {
	return path.Join(m.BaseDirectory, torrentsDirectory)
}

// MoviesDirectory returns absolute path to movies directory
func (m Manager) MoviesDirectory() string {
	return path.Join(m.BaseDirectory, moviesDirectory)
}
