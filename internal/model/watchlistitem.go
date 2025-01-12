package model

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type TorrentItem struct {
	models.SearchTorrentsResult
	TorrentContent string
}

type WatchListItem struct {
	// Global ID of media
	ID string `bson:"_id,omitempty"`

	// MovieInfo for movies or TV series
	MovieInfo rms_library.MovieInfo

	// Torrents contain suitable torrents for media item
	Torrents []TorrentItem

	// Seasons contain torrents for different seasons
	Seasons map[uint][]TorrentItem
}
