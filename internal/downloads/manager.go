package downloads

import (
	"context"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"go-micro.dev/v4/logger"
)

// Manager is responsible for downloading and management torrents
type Manager struct {
	cli       rms_torrent.RmsTorrentService
	onlineCli rms_torrent.RmsTorrentService
	dm        DirectoryManager
	db        Database
}

type content struct {
	contentType media.ContentType
	id          string
	seasons     map[uint]struct{}
}

// NewManager creates a Manager instance
func NewManager(cli rms_torrent.RmsTorrentService, onlineCli rms_torrent.RmsTorrentService, db Database, dm DirectoryManager, waitTorrentReady bool) *Manager {
	return &Manager{
		cli:       cli,
		onlineCli: onlineCli,
		db:        db,
		dm:        dm,
	}
}

func (m *Manager) client(onlinePlayback bool) rms_torrent.RmsTorrentService {
	if onlinePlayback {
		return m.onlineCli
	}
	return m.cli
}

// Initialize loads content and builds downloads index
func (m *Manager) Initialize() error {
	// создаем директории для фильмов
	// if err = m.dm.CreateMoviesLayout(movies); err != nil {
	// 	return err
	// }

	return nil
}

func (m *Manager) Download(ctx context.Context, item *model.ListItem, category string, torrent []byte) error {
	cli := m.client(item.List == rms_library.List_WatchList)

	req := rms_torrent.DownloadRequest{
		What:        torrent,
		Description: item.Title,
		Category:    category,
	}

	// ставим в очередь на скачивание торрент
	resp, err := cli.Download(ctx, &req)
	if err != nil {
		return fmt.Errorf("add torrent failed: %w", err)
	}

	torrentRecord := model.TorrentRecord{
		ID:       resp.Id,
		Title:    resp.Title,
		Online:   item.List == rms_library.List_WatchList,
		Location: resp.Location,
	}
	item.Torrents = append(item.Torrents, torrentRecord)

	logger.Infof("Torrent added, id = %s, %d files", resp.Id, len(resp.Files))

	if err = m.db.UpdateContent(ctx, item.ID, item.Torrents); err != nil {
		_, _ = cli.RemoveTorrent(ctx, &rms_torrent.RemoveTorrentRequest{Id: resp.Id})
		return fmt.Errorf("update movie content failed: %s", err)
	}

	return nil
}

func (m *Manager) RemoveTorrents(ctx context.Context, item *model.ListItem) error {
	total := []model.TorrentRecord{}
	for _, t := range item.Torrents {
		cli := m.client(t.Online)
		if _, err := cli.RemoveTorrent(ctx, &rms_torrent.RemoveTorrentRequest{Id: t.ID}); err != nil {
			logger.Warnf("Remove torrent failed: %s", err)
			total = append(total, t)
		}
	}

}
