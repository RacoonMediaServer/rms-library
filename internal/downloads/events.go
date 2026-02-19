package downloads

import (
	"context"
	"errors"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"go-micro.dev/v4/logger"
)

const addTimeout = 15 * time.Second

type eventNew struct {
	movie *model.Movie
}

type eventAdd struct {
	id       model.ID
	torrents []model.TorrentRecord
}

type eventRemove struct {
	id       model.ID
	torrents []model.TorrentRecord
}

type eventUpdate struct {
	id       model.ID
	torrents []model.TorrentRecord
}

func (m *Manager) processEvents() {
	for {
		e := <-m.eventChan
		switch event := e.(type) {
		case *eventNew:
			m.processEventNew(event)
		case *eventAdd:
			m.processEventAdd(event)
		case *eventUpdate:
			m.processEventUpdate(event)
		case *eventRemove:
			m.processEventRemove(event)
		default:
			logger.Fatalf("unknown event: %T", event)
		}
	}
}

func (m *Manager) getMovieInfo(id model.ID) (*rms_library.MovieInfo, error) {
	m.mu.Lock()
	mi, ok := m.movInfo[id]
	if ok {
		m.mu.Unlock()
		return mi, nil
	}
	m.mu.Unlock()

	mov, err := m.db.GetMovie(context.Background(), id)
	if err != nil {
		return nil, err
	}

	if mov == nil {
		return nil, errors.New("not found")
	}

	m.mu.Lock()
	m.movInfo[id] = &mov.Info
	m.mu.Unlock()

	return &mov.Info, nil
}

func (m *Manager) processEventNew(e *eventNew) {
	m.mu.Lock()
	m.movInfo[e.movie.ID] = &e.movie.Info
	m.mu.Unlock()

	// данные на диске появляются не сразу после добавления торрента, костыль
	<-time.After(addTimeout)

	m.createMovieLayout(e.movie)
}

func (m *Manager) processEventAdd(e *eventAdd) {
	mi, err := m.getMovieInfo(e.id)
	if err != nil {
		logger.Errorf("Create layout for new torrents failed [ %s ]: %s", e.id, err)
		return
	}

	// данные на диске появляются не сразу после добавления торрента, костыль
	<-time.After(addTimeout)

	for _, t := range e.torrents {
		m.layoutAddTorrent(mi, &t)
	}
}

func (m *Manager) processEventUpdate(e *eventUpdate) {
	mi, err := m.getMovieInfo(e.id)
	if err != nil {
		logger.Errorf("Update layout failed [ %s ]: %s", e.id, err)
		return
	}

	for _, t := range e.torrents {
		m.layoutRemoveTorrent(mi, &t)
		m.layoutAddTorrent(mi, &t)
	}
}

func (m *Manager) processEventRemove(e *eventRemove) {
	mi, err := m.getMovieInfo(e.id)
	if err != nil {
		logger.Errorf("Remove layout for new torrents failed [ %s ]: %s", e.id, err)
		return
	}

	for _, t := range e.torrents {
		m.layoutRemoveTorrent(mi, &t)
	}
}

func (m *Manager) createMovieLayout(mov *model.Movie) {
	for _, t := range mov.Torrents {
		m.layoutAddTorrent(&mov.Info, &t)
	}
}

func (m *Manager) layoutAddTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord) {
	if !t.Online {
		info, err := m.cli.GetTorrentInfo(context.Background(), &rms_torrent.GetTorrentInfoRequest{Id: t.ID})
		if err == nil {
			if info.Status != rms_torrent.Status_Done {
				return
			}
		} else {
			logger.Errorf("Get torrent status for '%s' failed: %s", t.ID, err)
		}
	}
	m.dm.MoviesMountTorrent(mi, t)
}

func (m *Manager) layoutRemoveTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord) {
	m.dm.MoviesUmountTorrent(mi, t)
}
