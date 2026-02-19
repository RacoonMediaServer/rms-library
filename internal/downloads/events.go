package downloads

import (
	"context"
	"errors"
	"time"

	"github.com/RacoonMediaServer/rms-library/v3/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
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
	for _, t := range e.movie.Torrents {
		m.mapTorrentToMedia[t.ID] = e.movie.ID
	}
	m.mu.Unlock()

	retry := []model.TorrentRecord{}
	for _, t := range e.movie.Torrents {
		if err := m.layoutAddTorrent(&e.movie.Info, &t); err != nil {
			retry = append(retry, t)
		}
	}

	if len(retry) != 0 {
		<-time.After(addTimeout)
		m.eventChan <- &eventAdd{id: e.movie.ID, torrents: retry}
	}
}

func (m *Manager) processEventAdd(e *eventAdd) {
	mi, err := m.getMovieInfo(e.id)
	if err != nil {
		logger.Errorf("Create layout for new torrents failed [ %s ]: %s", e.id, err)
		return
	}

	// данные на диске появляются не сразу после добавления торрента, костыль
	<-time.After(addTimeout)

	m.mu.Lock()
	for _, t := range e.torrents {
		m.mapTorrentToMedia[t.ID] = e.id
	}
	m.mu.Unlock()

	retry := []model.TorrentRecord{}
	for _, t := range e.torrents {
		if err := m.layoutAddTorrent(mi, &t); err != nil {
			retry = append(retry, t)
		}
	}

	if len(retry) != 0 {
		<-time.After(addTimeout)
		m.eventChan <- &eventAdd{id: e.id, torrents: retry}
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

func (m *Manager) layoutAddTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord) error {
	if !t.Online {
		info, err := m.cli.GetTorrentInfo(context.Background(), &rms_torrent.GetTorrentInfoRequest{Id: t.ID})
		if err == nil {
			if info.Status != rms_torrent.Status_Done {
				return nil
			}
		} else {
			logger.Errorf("Get torrent status for '%s' failed: %s", t.ID, err)
		}
	}

	return m.dm.MoviesMountTorrent(mi, t)
}

func (m *Manager) layoutRemoveTorrent(mi *rms_library.MovieInfo, t *model.TorrentRecord) {
	m.dm.MoviesUmountTorrent(mi, t)
}

func (m *Manager) handleExternalNotifications(ctx context.Context, event events.Notification) error {
	if event.TorrentID == nil {
		return nil
	}

	if event.Kind != events.Notification_DownloadComplete {
		return nil
	}

	m.mu.Lock()
	id, ok := m.mapTorrentToMedia[*event.TorrentID]
	m.mu.Unlock()
	if !ok {
		logger.Warnf("Info about torrent '%s' not found in the cached", event.TorrentID)
		return nil
	}

	info, err := m.cli.GetTorrentInfo(ctx, &rms_torrent.GetTorrentInfoRequest{Id: *event.TorrentID})
	if err != nil {
		logger.Errorf("Get torrent info failed: %s", err)
		return nil
	}

	m.eventChan <- &eventAdd{
		id: id,
		torrents: []model.TorrentRecord{
			{
				ID:       *event.TorrentID,
				Location: info.Location,
				Title:    info.Title,
				Size:     info.SizeMB,
			},
		},
	}

	return nil
}
