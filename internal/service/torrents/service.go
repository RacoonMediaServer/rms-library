package torrents

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RacoonMediaServer/rms-library/v3/internal/lock"
	"github.com/RacoonMediaServer/rms-library/v3/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	Locker    lock.Locker
	Database  Database
	Downloads DownloadsManager
	Movies    Movies
}

const lockTimeout = 15 * time.Second

func (s *Service) getItem(ctx context.Context, id model.ID) (*model.ListItem, lock.Unlocker, error) {
	l, err := lock.TimedLock(ctx, s.Locker, id, lockTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("acquire lock failed: %w", err)
	}

	item, err := s.Database.GetListItem(ctx, id)
	if err != nil {
		l.Unlock()
		return nil, nil, fmt.Errorf("get item from db failed: %w", err)
	}

	if item == nil {
		l.Unlock()
		return nil, nil, errors.New("item not found")
	}

	if item.List == rms_library.List_Archive {
		l.Unlock()
		return nil, nil, errors.New("item is archived")
	}

	return item, l, nil
}

// Add implements rms_library.TorrentsHandler.
func (s *Service) Add(ctx context.Context, req *rms_library.TorrentsAddRequest, resp *emptypb.Empty) error {
	if req.Link == nil && len(req.TorrentFile) == 0 {
		return errors.New("no torrent content presented")
	}
	if req.Link != nil && len(req.TorrentFile) != 0 {
		return errors.New("ambigous content presented")
	}

	id := model.ID(req.Id)

	item, lk, err := s.getItem(ctx, id)
	if err != nil {
		logger.Errorf("Get item %s failed: %s", id, err)
		return err
	}
	defer lk.Unlock()

	if err = s.download(ctx, item, req.Link, req.TorrentFile); err != nil {
		logger.Errorf("Add torrent to '%s' [ %s ] failed: %s", item.Title, item.ID, err)
		return err
	}

	logger.Infof("Torrent added to '%s' [ %s ]", item.Title, item.ID)
	return nil
}

// Delete implements rms_library.TorrentsHandler.
func (s *Service) Delete(ctx context.Context, req *rms_library.TorrentsDeleteRequest, resp *emptypb.Empty) error {
	id := model.ID(req.Id)

	item, lk, err := s.getItem(ctx, id)
	if err != nil {
		logger.Errorf("Get item %s failed: %s", id, err)
		return err
	}
	defer lk.Unlock()

	if err = s.Downloads.RemoveTorrent(ctx, item, req.TorrentId); err != nil {
		logger.Errorf("Remove torrent %s of '%s' [ %s ] failed: %s", req.TorrentId, item.Title, item.ID, err)
		return err
	}

	logger.Infof("Torrent %s of '%s' [ %s ] removed", req.TorrentId, item.Title, item.ID)
	return nil
}

func getSeasonPtr(season *uint32) *uint {
	if season == nil {
		return nil
	}
	tmpVal := uint(*season)
	return &tmpVal
}

// Find implements rms_library.TorrentsHandler.
func (s *Service) Find(ctx context.Context, req *rms_library.TorrentsFindRequest, resp *rms_library.TorrentsFindResponse) error {
	id := model.ID(req.Id)
	var err error
	switch id.ContentType() {
	case rms_library.ContentType_TypeMovies:
		resp.Torrents, err = s.Movies.FindTorrents(ctx, id, getSeasonPtr(req.Season))
	default:
		err = errors.New("unsupported content type")
	}
	if err != nil {
		logger.Errorf("Find torrents for %s failed: %s", req.Id, err)
	}
	return err
}

// List implements rms_library.TorrentsHandler.
func (s *Service) List(ctx context.Context, req *rms_library.TorrentsListRequest, resp *rms_library.TorrentsListResponse) error {
	id := model.ID(req.Id)

	item, err := s.Database.GetListItem(ctx, id)
	if err != nil {
		logger.Errorf("Get item %s failed: %s", id, err)
		return err
	}
	resp.Torrents = make([]*rms_library.Torrent, 0, len(item.Torrents))
	for _, t := range item.Torrents {
		tConverted := rms_library.Torrent{
			Id:    t.ID,
			Title: t.Title,
			Size:  t.Size,
		}
		resp.Torrents = append(resp.Torrents, &tConverted)
	}

	return nil
}
