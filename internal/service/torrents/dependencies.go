package torrents

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type Database interface {
	GetListItem(ctx context.Context, id model.ID) (*model.ListItem, error)
}

type DownloadsManager interface {
	Download(ctx context.Context, item *model.ListItem, torrent []byte) error
	RemoveTorrent(ctx context.Context, item *model.ListItem, torrentId string) error
}

type Movies interface {
	GetTorrentContent(ctx context.Context, torrentId string) ([]byte, error)
	FindTorrents(ctx context.Context, id model.ID, torrentId *string) ([]*rms_library.Torrent, error)
}
