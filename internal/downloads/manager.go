package downloads

import (
	"context"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"sync"
)

// Manager is responsible for mapping torrents to movies and storage layout
type Manager struct {
	mu               sync.RWMutex
	db               Database
	dm               DirectoryManager
	torrentToContent map[string]*content
	contentToTorrent map[string][]string
}

type content struct {
	contentType media.ContentType
	id          string
	seasons     map[uint]struct{}
}

// NewManager creates a Manager instance
func NewManager(db Database, dm DirectoryManager) *Manager {
	return &Manager{
		db:               db,
		dm:               dm,
		torrentToContent: map[string]*content{},
		contentToTorrent: map[string][]string{},
	}
}

// Initialize loads content and builds downloads index
func (m *Manager) Initialize() error {
	movies, err := m.db.SearchMovies(context.Background(), nil)
	if err != nil {
		return err
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

func (m *Manager) removeTorrent(torrentID string) {
	c := m.torrentToContent[torrentID]
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
		return
	}
	m.contentToTorrent[c.id] = tmp
}

// AddMovieDownload adds movie torrent to index, modify mov and returns torrents which need to delete
func (m *Manager) AddMovieDownload(torrentID string, mov *model.Movie, seasons map[uint]struct{}) (removeTorrents []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// для фильмов заменяем торрент и удаляем старый
	if mov.Info.Type == rms_library.MovieType_Film {
		if mov.TorrentID != "" {
			removeTorrents = append(removeTorrents, mov.TorrentID)
			m.removeTorrent(mov.TorrentID)
		}

		mov.ReplaceTorrentID(torrentID)

		m.torrentToContent[torrentID] = &content{
			contentType: media.Movies,
			id:          mov.ID,
		}
		m.contentToTorrent[mov.ID] = []string{torrentID}
		return
	}

	// ищем, какие сезоны заменяются с новым скачиванием
	oldTorrents := mov.AddOrReplaceSeasons(torrentID, seasons)

	// удаляем торренты, которые нигде больше не используются
	for t, s := range oldTorrents {
		c := m.torrentToContent[t]
		if c.seasons != nil {
			for _, no := range s {
				delete(c.seasons, no)
			}
		}
		if len(c.seasons) == 0 {
			removeTorrents = append(removeTorrents, t)
			m.removeTorrent(t)
		}
	}

	// добавляем торрент
	m.torrentToContent[torrentID] = &content{
		contentType: media.Movies,
		id:          mov.ID,
		seasons:     seasons,
	}
	tmp := m.contentToTorrent[mov.ID]
	tmp = append(tmp, torrentID)
	m.contentToTorrent[mov.ID] = tmp

	return
}

func (m *Manager) RemoveMovie(mov *model.Movie) (removeTorrents []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	torrents, ok := m.contentToTorrent[mov.ID]
	if !ok {
		return nil
	}
	for _, t := range torrents {
		m.removeTorrent(t)
		removeTorrents = append(removeTorrents, t)
	}
	_ = m.dm.DeleteMovieLayout(mov)
	return torrents
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

func (m *Manager) RemoveMovieTorrent(torrentID string, mov *model.Movie) (removeMovie bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.torrentToContent[torrentID]
	if !ok {
		return false
	}

	for no, _ := range c.seasons {
		mov.RemoveSeason(no)
	}
	m.removeTorrent(torrentID)

	if len(mov.Seasons) == 0 {
		_ = m.dm.DeleteMovieLayout(mov)
		return true
	} else {
		_ = m.dm.CreateMovieLayout(mov)
	}

	return false
}
