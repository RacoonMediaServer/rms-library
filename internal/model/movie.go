package model

import rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"

// Season represents one season of TV series
type Season struct {
	// No is a number of season
	No uint

	// TorrentID is a torrent ID on rms-torrent service
	TorrentID string

	// Episodes is a list of contained media
	Episodes []File
}

// Movie represents info about downloaded movie
type Movie struct {
	// Global ID of movie or series (related to imdb.com)
	ID string `bson:"_id,omitempty"`

	// Info about movie/series
	Info rms_library.MovieInfo

	// Files maps torrent IDs to file set on data storage (for films)
	Files map[string][]File

	// Seasons contain all info about downloaded seasons of TV series
	Seasons []Season
}

func (m *Movie) IsSeasonDownloaded(no uint) bool {
	for i := range m.Seasons {
		if m.Seasons[i].No == no {
			return true
		}
	}
	return false
}

func (m *Movie) AddFile(torrentID string, f File, season uint) {
	if m.Info.Type == rms_library.MovieType_Film || season == 0 {
		if m.Files == nil {
			m.Files = make(map[string][]File)
		}
		m.Files[torrentID] = append(m.Files[torrentID], f)
		return
	}

	for i := range m.Seasons {
		s := &m.Seasons[i]
		if s.TorrentID == torrentID && s.No == season {
			s.Episodes = append(s.Episodes, f)
			return
		}
	}

	s := Season{
		No:        season,
		TorrentID: torrentID,
		Episodes:  []File{f},
	}
	m.Seasons = append(m.Seasons, s)
}
