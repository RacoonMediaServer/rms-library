package downloads

import (
	"context"
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/analysis"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"go-micro.dev/v4/logger"
	"sync"
)

// Manager is responsible for downloading and management torrents
type Manager struct {
	mu sync.RWMutex

	cli              rms_torrent.RmsTorrentService
	dm               DirectoryManager
	db               Database
	torrentToContent map[string]*content
	contentToTorrent map[string][]string
}

type content struct {
	contentType media.ContentType
	id          string
	seasons     map[uint]struct{}
}

// NewManager creates a Manager instance
func NewManager(cli rms_torrent.RmsTorrentService, db Database, dm DirectoryManager) *Manager {
	return &Manager{
		cli:              cli,
		db:               db,
		dm:               dm,
		torrentToContent: map[string]*content{},
		contentToTorrent: map[string][]string{},
	}
}

// Initialize loads content and builds downloads index
func (m *Manager) Initialize() error {
	// создаем раскладку директорий
	if err := m.dm.CreateDefaultLayout(); err != nil {
		return fmt.Errorf("create default layout failed: %s", err)
	}

	// загружаем все фильмы
	movies, err := m.db.SearchMovies(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("get movies from database failed: %s", err)
	}

	// заполняем индекс фильмами
	for _, mov := range movies {
		if mov.TorrentID == "" {
			for no, season := range mov.Seasons {
				c, ok := m.torrentToContent[season.TorrentID]
				if !ok {
					c = &content{
						contentType: media.Movies,
						id:          mov.ID,
						seasons:     map[uint]struct{}{},
					}
					m.torrentToContent[season.TorrentID] = c
				}
				c.seasons[no] = struct{}{}

				tmp := m.contentToTorrent[mov.ID]
				tmp = append(tmp, season.TorrentID)
				m.contentToTorrent[mov.ID] = tmp
			}
		} else {
			m.torrentToContent[mov.TorrentID] = &content{
				contentType: media.Movies,
				id:          mov.ID,
			}
			m.contentToTorrent[mov.ID] = []string{mov.TorrentID}
		}
	}

	// создаем директории для фильмов
	if err = m.dm.CreateMoviesLayout(movies); err != nil {
		return err
	}

	return nil
}

func (m *Manager) removeTorrent(torrentID string, onlyFromCache bool) {
	c, ok := m.torrentToContent[torrentID]
	if ok {
		delete(m.torrentToContent, torrentID)
		tmp := m.contentToTorrent[c.id]
		for i, t := range tmp {
			if t == torrentID {
				tmp[i] = tmp[len(tmp)-1]
				tmp = tmp[:len(tmp)-1]
				break
			}
		}
		if len(tmp) == 0 {
			delete(m.contentToTorrent, c.id)
		} else {
			m.contentToTorrent[c.id] = tmp
		}
	}
	if !onlyFromCache {
		if _, err := m.cli.RemoveTorrent(context.Background(), &rms_torrent.RemoveTorrentRequest{Id: torrentID}); err != nil {
			logger.Errorf("Delete torrent %s failed: %s", torrentID, err)
		}
	}
}

func getUniqueSeasons(results []analysis.Result) map[uint]struct{} {
	m := map[uint]struct{}{}
	for _, r := range results {
		if r.Season != 0 {
			m[r.Season] = struct{}{}
		}
	}
	return m
}

// DownloadMovie adds torrent to download and update movie info
func (m *Manager) DownloadMovie(ctx context.Context, mov *model.Movie, torrent []byte) error {
	var torrentsToDelete []string

	// ставим в очередь на скачивание торрент
	resp, err := m.cli.Download(ctx, &rms_torrent.DownloadRequest{What: torrent})
	if err != nil {
		return fmt.Errorf("add torrent failed: %w", err)
	}

	// анализируем контент раздачи
	var results []analysis.Result
	for _, file := range resp.Files {
		results = append(results, analysis.Analyze(file))
	}

	// какие то сезоны необходимо заменить новыми
	seasons := getUniqueSeasons(results)
	oldTorrents := mov.AddOrReplaceSeasons(resp.Id, seasons)

	// помечаем прошлый торрент на удаление если происходит замена раздачи фильма
	if mov.Info.Type == rms_library.MovieType_Film && mov.TorrentID != "" {
		torrentsToDelete = append(torrentsToDelete, mov.TorrentID)
		mov.ReplaceTorrentID(resp.Id)
	}

	// накидываем файлы
	for i, file := range resp.Files {
		f := model.File{
			Path:  file,
			Title: results[i].EpisodeName,
			Type:  results[i].FileType,
			No:    results[i].Episode,
		}
		mov.AddFile(resp.Id, f, results[i].Season)
	}

	if err = m.db.UpdateMovieContent(ctx, mov); err != nil {
		m.removeTorrent(resp.Id, false)
		return fmt.Errorf("update movie content failed: %s", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// удаляем торренты, которые нигде больше не используются
	for t, s := range oldTorrents {
		c := m.torrentToContent[t]
		if c.seasons != nil {
			for _, no := range s {
				delete(c.seasons, no)
			}
		}
		if len(c.seasons) == 0 {
			torrentsToDelete = append(torrentsToDelete, t)
		}
	}

	// добавляем торрент
	m.torrentToContent[resp.Id] = &content{
		contentType: media.Movies,
		id:          mov.ID,
		seasons:     seasons,
	}
	tmp := m.contentToTorrent[mov.ID]
	tmp = append(tmp, resp.Id)
	m.contentToTorrent[mov.ID] = tmp

	for _, t := range torrentsToDelete {
		m.removeTorrent(t, false)
	}

	return nil
}

func (m *Manager) RemoveMovie(ctx context.Context, mov *model.Movie) error {
	if err := m.db.DeleteMovie(ctx, mov.ID); err != nil {
		return fmt.Errorf("delete movie from database failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	torrents, ok := m.contentToTorrent[mov.ID]
	if !ok {
		return nil
	}
	for _, t := range torrents {
		m.removeTorrent(t, false)
	}
	if err := m.dm.DeleteMovieLayout(mov); err != nil {
		logger.Errorf("Delete movie '%s' layout failed: %s", mov.Info.Title, err)
	}
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
		logger.Infof("Movie '%s' download complete. creating layout", mov.Info.Title)
		if err := m.dm.CreateMovieLayout(mov); err != nil {
			logger.Errorf("Create layout for movie '%s' failed: %s", err)
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

	c, ok := m.torrentToContent[torrentID]
	if !ok {
		return
	}

	for no, _ := range c.seasons {
		mov.RemoveSeason(no)
	}

	if len(mov.Seasons) != 0 {
		if err := m.db.UpdateMovieContent(context.Background(), mov); err != nil {
			logger.Errorf("Update movie '%s' in database failed: %s", mov.Info.Title, err)
			return
		}
		if err := m.dm.CreateMovieLayout(mov); err != nil {
			logger.Errorf("Update layout for movie '%s' failed: %s", mov.Info.Title, err)
		}
		m.removeTorrent(torrentID, true)
		return
	}

	if err := m.db.DeleteMovie(context.Background(), mov.ID); err != nil {
		logger.Errorf("Delete movie '%s' from database failed: %s", mov.Info.Title, err)
		return
	}

	m.removeTorrent(torrentID, true)

	if err := m.dm.DeleteMovieLayout(mov); err != nil {
		logger.Errorf("Delete layout for movie '%s' failed: %s", mov.Info.Title, err)
	}
}
