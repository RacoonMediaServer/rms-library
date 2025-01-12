package model

import (
	"time"

	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

// Movie represents info about downloaded movie
type Movie struct {
	// Global ID of movie or series (related to themoviedb.org)
	ID string `bson:"_id,omitempty"`

	// Info about movie/series
	Info rms_library.MovieInfo

	// ID of associated torrents
	Torrents []string

	// LastAvailableCheck is a time when check of available new seasons has been occurred
	LastAvailableCheck time.Time

	// AvailableSeasons contains season which available on trackers
	AvailableSeasons []uint

	// Seasons contain all info about downloaded seasons of TV series
	Seasons map[uint]bool

	// Voice contains downloaded voice for series seasons
	Voice string
}

func (m *Movie) IsSeasonDownloaded(no uint) bool {
	return m.Seasons[no]
}

func (m *Movie) SetVoice(voice string) {
	if m.Voice == "" {
		m.Voice = voice
	}
}

func (m *Movie) RemoveSeason(no uint) {
	m.Seasons[no] = false
}
