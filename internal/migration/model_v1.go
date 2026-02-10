package migration

import (
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type movieV1 struct {
	// Global ID of movie or series (related to themoviedb.org)
	ID string `bson:"_id,omitempty"`

	// Info about movie/series
	Info rms_library.MovieInfo

	// ID of associated torrents
	Torrents []model.TorrentRecord

	// LastAvailableCheck is a time when check of available new seasons has been occurred
	LastAvailableCheck time.Time

	// AvailableSeasons contains season which available on trackers
	AvailableSeasons []uint

	// Voice contains downloaded voice for series seasons
	Voice string
}

type torrentItemV1 struct {
	models.SearchTorrentsResult
	TorrentContent string
}

type watchListItemV1 struct {
	// Global ID of media
	ID string `bson:"_id,omitempty"`

	// MovieInfo for movies or TV series
	MovieInfo rms_library.MovieInfo

	// Torrents contain suitable torrents for media item
	Torrents []torrentItemV1

	// Seasons contain torrents for different seasons
	Seasons map[uint][]torrentItemV1
}
