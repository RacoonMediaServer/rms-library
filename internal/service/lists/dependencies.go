package lists

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type Database interface {
	GetListItems(ctx context.Context, list rms_library.List, contentType *rms_library.ContentType) (results []*model.ListItem, err error)
	MoveListItem(ctx context.Context, id model.ID, newList rms_library.List) error
	GetListItem(ctx context.Context, id model.ID) (*model.ListItem, error)
	DeleteListItem(ctx context.Context, id model.ID) error
}

type Movies interface {
	Add(ctx context.Context, id model.ID, list rms_library.List) error
}

type Scheduler interface {
	Cancel(group string)
}

type DownloadManager interface {
	DropTorrents(ctx context.Context, torrents []model.TorrentRecord)
}
