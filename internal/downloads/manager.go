package downloads

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
)

// Manager is responsible for downloading and management torrents
type Manager struct {
	mu sync.RWMutex

	cli              rms_torrent.RmsTorrentService
	onlineCli        rms_torrent.RmsTorrentService
	dm               DirectoryManager
	db               Database
	torrentToContent map[string]*content
	waitTorrentReady bool
}

type content struct {
	contentType media.ContentType
	id          string
	seasons     map[uint]struct{}
}

// NewManager creates a Manager instance
func NewManager(cli rms_torrent.RmsTorrentService, onlineCli rms_torrent.RmsTorrentService, db Database, dm DirectoryManager, waitTorrentReady bool) *Manager {
	return &Manager{
		cli:              cli,
		onlineCli:        onlineCli,
		db:               db,
		dm:               dm,
		torrentToContent: map[string]*content{},
		waitTorrentReady: waitTorrentReady,
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
	// загружаем все фильмы
	movies, err := m.db.SearchMovies(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("get movies from database failed: %s", err)
	}

	// заполняем индекс фильмами
	for _, mov := range movies {
		for _, t := range mov.Torrents {
			m.torrentToContent[t.ID] = &content{
				contentType: media.Movies,
				id:          mov.ID,
			}
		}
	}

	// создаем директории для фильмов
	if err = m.dm.CreateMoviesLayout(movies); err != nil {
		return err
	}

	return nil
}

func (m *Manager) removeTorrent(t *model.TorrentRecord, onlyFromCache bool) {
	delete(m.torrentToContent, t.ID)

	if !onlyFromCache {
		cli := m.client(t.Online)
		if _, err := cli.RemoveTorrent(context.Background(), &rms_torrent.RemoveTorrentRequest{Id: t.ID}); err != nil {
			logger.Errorf("Delete torrent %s failed: %s", t.ID, err)
		}
	}
}

// DownloadMovie adds torrent to download and update movie info
func (m *Manager) DownloadMovie(ctx context.Context, mov *model.Movie, voice string, torrent []byte, online bool) error {
	cli := m.client(online)
	req := rms_torrent.DownloadRequest{
		What:        torrent,
		Description: mov.Info.Title,
		Category:    model.GetCategory(mov.Info.Type),
	}

	// ставим в очередь на скачивание торрент
	resp, err := cli.Download(ctx, &req)
	if err != nil {
		return fmt.Errorf("add torrent failed: %w", err)
	}

	torrentRecord := model.TorrentRecord{
		ID:       resp.Id,
		Title:    resp.Title,
		Online:   online,
		Location: resp.Location,
	}
	mov.AddTorrent(torrentRecord)

	logger.Infof("Torrent added, id = %s, %d files", resp.Id, len(resp.Files))

	mov.SetVoice(voice)

	if err = m.db.UpdateMovieContent(ctx, mov); err != nil {
		m.removeTorrent(&torrentRecord, false)
		return fmt.Errorf("update movie content failed: %s", err)
	}

	if !m.waitTorrentReady || online {
		defer m.dm.CreateMovieLayout(mov)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// добавляем торрент
	m.torrentToContent[resp.Id] = &content{
		contentType: media.Movies,
		id:          mov.ID,
	}

	return nil
}

func (m *Manager) RemoveMovie(ctx context.Context, mov *model.Movie) error {
	if err := m.db.DeleteMovie(ctx, mov.ID); err != nil {
		return fmt.Errorf("delete movie from database failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range mov.Torrents {
		m.removeTorrent(&t, false)
	}
	m.dm.DeleteMovieLayout(mov)
	return nil
}

func (m *Manager) GetMovieByTorrent(torrentID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.torrentToContent[torrentID]
	if !ok {
		return "", ok
	}
	return c.id, ok
}

func (m *Manager) HandleTorrentEvent(kind events.Notification_Kind, torrentID string, mov *model.Movie) {
	switch kind {
	case events.Notification_DownloadComplete:
		if m.waitTorrentReady {
			m.torrentIsReady(torrentID, mov)
		}

	case events.Notification_TorrentRemoved:
		logger.Infof("Torrent %s of movie '%s' removed", torrentID, mov.Info.Title)
		m.removeMovieTorrent(torrentID, mov)
	default:
		return
	}
}

func (m *Manager) removeMovieTorrent(torrentID string, mov *model.Movie) {
	m.mu.Lock()
	defer m.mu.Unlock()

	torrentRecord, ok := mov.RemoveTorrent(torrentID)

	if !ok {
		return
	}

	if len(mov.Torrents) != 0 {
		if err := m.db.UpdateMovieContent(context.Background(), mov); err != nil {
			logger.Errorf("Update movie '%s' in database failed: %s", mov.Info.Title, err)
			return
		}
		m.dm.CreateMovieLayout(mov)
		m.removeTorrent(&torrentRecord, true)
		return
	}

	if err := m.db.DeleteMovie(context.Background(), mov.ID); err != nil {
		logger.Errorf("Delete movie '%s' from database failed: %s", mov.Info.Title, err)
		return
	}

	m.removeTorrent(&torrentRecord, true)

	m.dm.DeleteMovieLayout(mov)
}

func (m *Manager) GetMovieStoreSize(ctx context.Context, mov *model.Movie) uint64 {
	var size uint64
	for _, t := range mov.Torrents {
		if t.Online {
			continue
		}
		info, err := m.cli.GetTorrentInfo(ctx, &rms_torrent.GetTorrentInfoRequest{Id: t.ID}, client.WithRequestTimeout(5*time.Second))
		if err != nil {
			logger.Warnf("Get torrent info failed: %s", err)
			continue
		}
		if info.Status == rms_torrent.Status_Done {
			size += info.SizeMB
		}
	}
	return size
}

func (m *Manager) torrentIsReady(torrentId string, mov *model.Movie) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := m.cli.GetTorrentInfo(ctx, &rms_torrent.GetTorrentInfoRequest{Id: torrentId})
	if err == nil {
		tor := mov.GetTorrent(torrentId)
		if tor != nil {
			tor.Location = info.Location
			if err = m.db.UpdateMovieContent(ctx, mov); err != nil {
				logger.Warnf("Update torrent location failed: %s", err)
			}
		}
	}

	logger.Infof("Movie '%s' download complete. creating layout", mov.Info.Title)
	m.dm.CreateMovieLayout(mov)
}
