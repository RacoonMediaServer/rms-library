package storage

import (
	"path"
)

const torrentsDirectory = "torrents"
const moviesDirectory = "movies"
const downloadsDirectory = "downloads"

// TorrentsDirectory returns absolute path to torrents directory
func (m *Manager) TorrentsDirectory() string {
	return path.Join(m.base, torrentsDirectory)
}

// MoviesDirectory returns absolute path to movies directory
func (m *Manager) MoviesDirectory() string {
	return path.Join(m.base, moviesDirectory)
}

// DownloadsDirectory returns absolute path to directory with different downloaded from torrents content
func (m *Manager) DownloadsDirectory() string {
	return path.Join(m.base, downloadsDirectory)
}
