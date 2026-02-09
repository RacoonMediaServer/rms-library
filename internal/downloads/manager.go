package downloads

import (
	"context"
	"errors"
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

func (m *Manager) Download(ctx context.Context, item *model.ListItem, torrent []byte) error {
	cli := m.client(item.List == rms_library.List_WatchList)

	req := rms_torrent.DownloadRequest{
		What:        torrent,
		Description: item.Title,
		Category:    item.Category,
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

func (m *Manager) DropTorrents(ctx context.Context, torrents []model.TorrentRecord) {
	for _, t := range torrents {
		cli := m.client(t.Online)
		if _, err := cli.RemoveTorrent(ctx, &rms_torrent.RemoveTorrentRequest{Id: t.ID}); err != nil {
			logger.Warnf("Remove torrent failed: %s", err)
		}
	}
}

func (m *Manager) RemoveTorrent(ctx context.Context, item *model.ListItem, torrentId string) error {
	var target *model.TorrentRecord
	var updatedTorrents []model.TorrentRecord
	for i := range item.Torrents {
		if item.Torrents[i].ID == torrentId {
			target = &item.Torrents[i]
			updatedTorrents = append(item.Torrents[:i], item.Torrents[i+1:]...)
			break
		}
	}

	if target == nil {
		return errors.New("torrent not found")
	}

	cli := m.client(target.Online)
	if _, err := cli.RemoveTorrent(ctx, &rms_torrent.RemoveTorrentRequest{Id: target.ID}); err != nil {
		return err
	}

	if err := m.db.UpdateContent(ctx, item.ID, updatedTorrents); err != nil {
		return err
	}

	item.Torrents = updatedTorrents
	return nil
}

func (m *Manager) getTorrentsMap(ctx context.Context, online bool) (map[string]*rms_torrent.TorrentInfo, error) {
	cli := m.client(online)

	resp, err := cli.GetTorrents(ctx, &rms_torrent.GetTorrentsRequest{IncludeDoneTorrents: true})
	if err != nil {
		return nil, err
	}

	result := map[string]*rms_torrent.TorrentInfo{}
	for _, t := range resp.Torrents {
		result[t.Id] = t
	}

	return result, nil
}

func (m *Manager) DropMissedTorrents(ctx context.Context, item *model.ListItem) error {
	offlineTorrents, err := m.getTorrentsMap(ctx, false)
	if err != nil {
		return err
	}

	onlineTorrents, err := m.getTorrentsMap(ctx, true)
	if err != nil {
		return err
	}

	resultTorrents := make([]model.TorrentRecord, 0, len(item.Torrents))
	changed := false
	for _, t := range item.Torrents {
		torrents := offlineTorrents
		if t.Online {
			torrents = onlineTorrents
		}
		_, found := torrents[t.ID]
		if !found {
			changed = true
		} else {
			resultTorrents = append(resultTorrents, t)
		}
	}

	if !changed {
		return nil
	}

	if err = m.db.UpdateContent(ctx, item.ID, resultTorrents); err != nil {
		return err
	}

	item.Torrents = resultTorrents
	return nil
}

func (m *Manager) UpdateTorrentInfo(ctx context.Context, item *model.ListItem) error {
	changed := false
	for i := range item.Torrents {
		t := &item.Torrents[i]
		cli := m.client(t.Online)
		info, err := cli.GetTorrentInfo(ctx, &rms_torrent.GetTorrentInfoRequest{Id: t.ID})
		if err != nil {
			return fmt.Errorf("get torrent info about %s failed: %w", t.ID, err)
		}
		if info.Location != t.Location {
			t.Location = info.Location
			changed = true
		}
		if !t.Online && info.SizeMB != t.Size {
			t.Size = info.SizeMB
			changed = true
		}
	}

	if !changed {
		return nil
	}

	return m.db.UpdateContent(ctx, item.ID, item.Torrents)
}
