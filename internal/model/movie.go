package model

import rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"

// Season represents one season of TV series
type Season struct {
	// TorrentID is a torrent ID of season on rms-torrent service
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

	// ID of torrent of entire media content (can be empty)
	TorrentID string

	// Files contains of film files
	Files []File

	// Seasons contain all info about downloaded seasons of TV series
	Seasons map[uint]*Season
}

func (m *Movie) IsSeasonDownloaded(no uint) bool {
	_, ok := m.Seasons[no]
	return ok
}

func (m *Movie) AddFile(torrentID string, f File, season uint) {
	if m.Info.Type == rms_library.MovieType_Film || season == 0 {
		m.TorrentID = torrentID
		m.Files = append(m.Files, f)
		return
	}
	if m.Seasons == nil {
		m.Seasons = make(map[uint]*Season)
	}
	s, ok := m.Seasons[season]
	if !ok {
		s = &Season{
			TorrentID: torrentID,
			Episodes:  []File{f},
		}
		m.Seasons[season] = s
	}

	s.Episodes = append(s.Episodes, f)
}

func (m *Movie) ReplaceTorrentID(torrentID string) {
	m.TorrentID = torrentID
	m.Files = nil
}

func (m *Movie) AddOrReplaceSeasons(torrentID string, seasons map[uint]struct{}) (oldTorrents map[string][]uint) {
	var newSeasons []uint
	oldTorrents = map[string][]uint{}

	for no, _ := range seasons {
		season, ok := m.Seasons[no]
		if !ok {
			newSeasons = append(newSeasons, no)
			continue
		}
		oldTorrents[season.TorrentID] = append(oldTorrents[season.TorrentID], no)
		season.TorrentID = torrentID
		season.Episodes = nil
	}

	for _, no := range newSeasons {
		if m.Seasons == nil {
			m.Seasons = map[uint]*Season{}
		}
		m.Seasons[no] = &Season{
			TorrentID: torrentID,
		}
	}
	return
}

func (m *Movie) FindSeasonByTorrentID(torrentID string) (uint, bool) {
	for no, s := range m.Seasons {
		if s.TorrentID == torrentID {
			return no, true
		}
	}
	return 0, false
}

func (m *Movie) RemoveSeason(no uint) {
	delete(m.Seasons, no)
}
