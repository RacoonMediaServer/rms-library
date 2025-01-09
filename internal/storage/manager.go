package storage

import (
	"fmt"
	"os"

	"github.com/RacoonMediaServer/rms-library/internal/config"
)

const mediaPerms = 0755
const downloadPerms = 0777

const maxFsCommands = 50

// Manager is responsible for management content on a disk
type Manager struct {
	cmd  chan func()
	dirs config.Directories
}

// NewManager creates Manager and base directory layout
func NewManager(dirs config.Directories) (*Manager, error) {
	m := &Manager{dirs: dirs}

	if err := os.MkdirAll(dirs.Downloads, downloadPerms); err != nil {
		return nil, fmt.Errorf("create downloads directory failed: %w", err)
	}
	if err := os.MkdirAll(dirs.Content, mediaPerms); err != nil {
		return nil, fmt.Errorf("create content directory failed: %w", err)
	}
	if err := os.MkdirAll(dirs.WatchList, mediaPerms); err != nil {
		return nil, fmt.Errorf("create watchlist directory failed: %w", err)
	}
	m.cmd = make(chan func(), maxFsCommands)
	go func() {
		for cmd := range m.cmd {
			cmd()
		}
	}()

	return m, nil
}
