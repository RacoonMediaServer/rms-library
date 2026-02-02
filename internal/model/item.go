package model

import (
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type ListItem struct {
	// Global ID of movie or series (related to themoviedb.org)
	ID ID `bson:"_id,omitempty"`

	// Title is the item title
	Title string

	// List which movie belongs to
	List rms_library.List

	// ID of associated torrents
	Torrents []TorrentRecord
}

type TorrentRecord struct {
	ID       string
	Title    string
	Location string
	Size     uint64
	Online   bool
}
