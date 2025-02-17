package storage

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (m *Manager) StoreWatchListTorrent(itemTitle string, torrent []byte) (id string, err error) {
	itemTitle = escape(itemTitle)
	fileName := uuid.NewString() + ".torrent"
	id = filepath.Join(itemTitle, fileName)
	err = os.MkdirAll(filepath.Join(m.dirs.WatchList, itemTitle), mediaPerms)
	if err == nil {
		err = os.WriteFile(filepath.Join(m.dirs.WatchList, id), torrent, mediaPerms)
	}
	return
}

func (m *Manager) LoadWatchListTorrent(contentPath string) ([]byte, error) {
	return os.ReadFile(filepath.Join(m.dirs.WatchList, contentPath))
}
