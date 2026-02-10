package model

import (
	"time"

	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type ListItem struct {
	// Global ID of movie or series (related to themoviedb.org)
	ID ID `bson:"_id,omitempty"`

	// When item added
	CreatedAt time.Time

	// Title is the item title
	Title string

	// List which movie belongs to
	List rms_library.List

	// Category for downloads
	Category string

	// Media content type
	ContentType rms_library.ContentType

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

func (li *ListItem) Size() uint64 {
	var result uint64
	for _, t := range li.Torrents {
		if !t.Online {
			result += t.Size
		}
	}
	return result
}
