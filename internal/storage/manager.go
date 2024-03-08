package storage

import (
	"fmt"
	"os"
)

const mediaPerms = 0755
const downloadPerms = 0777

const maxFsCommands = 50

// Manager is responsible for management content on a disk
type Manager struct {
	base           string
	cmd            chan func()
	fixTorrentPath bool
}

// NewManager creates Manager and base directory layout
func NewManager(baseDirectory string, fixTorrentPath bool) (*Manager, error) {
	m := &Manager{base: baseDirectory, fixTorrentPath: fixTorrentPath}

	if err := os.MkdirAll(m.TorrentsDirectory(), 0777); err != nil {
		return nil, fmt.Errorf("create torrents directory failed: %w", err)
	}
	if err := os.MkdirAll(m.MoviesDirectory(), mediaPerms); err != nil {
		return nil, fmt.Errorf("create torrents directory failed: %w", err)
	}
	if err := os.MkdirAll(m.DownloadsDirectory(), downloadPerms); err != nil {
		return nil, fmt.Errorf("create downloads directory failed: %w", err)
	}
	m.cmd = make(chan func(), maxFsCommands)
	go func() {
		for cmd := range m.cmd {
			cmd()
		}
	}()

	return m, nil
}
