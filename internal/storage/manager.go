package storage

import (
	"fmt"
	"os"

	"github.com/RacoonMediaServer/rms-library/v3/internal/config"
)

const mediaPerms = 0755
const downloadPerms = 0777

const maxFsCommands = 50

// Manager is responsible for management content on a disk
type Manager struct {
	dirs config.Directories
}

// NewManager creates Manager and base directory layout
func NewManager(dirs config.Directories) (*Manager, error) {
	m := &Manager{
		dirs: dirs,
	}

	_ = os.RemoveAll(dirs.Content)

	if err := os.MkdirAll(dirs.Downloads, downloadPerms); err != nil {
		return nil, fmt.Errorf("create downloads directory failed: %w", err)
	}
	if err := os.MkdirAll(dirs.Content, mediaPerms); err != nil {
		return nil, fmt.Errorf("create content directory failed: %w", err)
	}
	if err := os.MkdirAll(dirs.WatchList, mediaPerms); err != nil {
		return nil, fmt.Errorf("create watchlist directory failed: %w", err)
	}

	return m, nil
}
