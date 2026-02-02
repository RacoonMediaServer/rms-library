package model

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type TorrentSearchResult struct {
	models.SearchMoviesResult
	ID string
}

// Movie represents info about downloaded movie
type Movie struct {
	ListItem `bson:",inline"`

	// Info about movie/series
	Info rms_library.MovieInfo

	// Voice contains downloaded voice for series seasons
	Voice string

	// ArchivedTorrents contains all search results
	ArchivedTorrents []TorrentSearchResult

	// ArchivedSeasons contains all seasons search results
	ArchivedSeasons map[uint]TorrentSearchResult
}

func (m *Movie) SetVoice(voice string) {
	if m.Voice == "" {
		m.Voice = voice
	}
}

func (m *Movie) RemoveTorrent(id string) (TorrentRecord, bool) {
	for i := range m.Torrents {
		if m.Torrents[i].ID == id {
			result := m.Torrents[i]
			m.Torrents = append(m.Torrents[:i], m.Torrents[i+1:]...)
			return result, true
		}
	}

	return TorrentRecord{}, false
}

func (m *Movie) AddTorrent(record TorrentRecord) {
	m.Torrents = append(m.Torrents, record)
}

func (m *Movie) GetTorrent(id string) *TorrentRecord {
	for i := range m.Torrents {
		if m.Torrents[i].ID == id {
			return &m.Torrents[i]
		}
	}
	return nil
}
