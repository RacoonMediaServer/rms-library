package model

import (
	"time"

	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type TorrentRecord struct {
	ID    string
	Title string
}

// Movie represents info about downloaded movie
type Movie struct {
	// Global ID of movie or series (related to themoviedb.org)
	ID string `bson:"_id,omitempty"`

	// Info about movie/series
	Info rms_library.MovieInfo

	// ID of associated torrents
	Torrents []TorrentRecord

	// LastAvailableCheck is a time when check of available new seasons has been occurred
	LastAvailableCheck time.Time

	// AvailableSeasons contains season which available on trackers
	AvailableSeasons []uint

	// Voice contains downloaded voice for series seasons
	Voice string
}

func (m *Movie) SetVoice(voice string) {
	if m.Voice == "" {
		m.Voice = voice
	}
}

func (m *Movie) RemoveTorrent(id string) bool {
	for i := range m.Torrents {
		if m.Torrents[i].ID == id {
			m.Torrents = append(m.Torrents[:i], m.Torrents[i+1:]...)
			return true
		}
	}

	return false
}

func (m *Movie) AddTorrent(id, title string) {
	m.Torrents = append(m.Torrents, TorrentRecord{ID: id, Title: title})
}
