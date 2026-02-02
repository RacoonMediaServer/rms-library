package lists

import (
	"context"
	"errors"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	Database  Database
	Movies    Movies
	Scheduler Scheduler
	Downloads DownloadManager
}

// Add implements rms_library.ListsHandler.
func (s *Service) Add(ctx context.Context, req *rms_library.ListsAddRequest, resp *emptypb.Empty) error {
	id := model.ID(req.Id)
	var err error
	switch id.ContentType() {
	case rms_library.ContentType_TypeMovies:
		err = s.Movies.Add(ctx, id, req.List)
	default:
		err = errors.New("unsupported content type")
	}

	if err != nil {
		logger.Errorf("Add content failed: %s", err)
	}
	return err
}

// Delete implements rms_library.ListsHandler.
func (s *Service) Delete(ctx context.Context, req *rms_library.ListsDeleteRequest, resp *emptypb.Empty) error {
	item, err := s.Database.GetListItem(ctx, model.ID(req.Id))
	if err != nil {
		logger.Errorf("Cannot get '%s' item", req.Id, err)
		return err
	}
	if item == nil {
		return errors.New("not found")
	}
	if err = s.Database.DeleteListItem(ctx, model.ID(req.Id)); err != nil {
		logger.Errorf("Delete '%s'from db failed: %s", req.Id, err)
		return err
	}

	s.Scheduler.Cancel(req.Id)
	s.Downloads.RemoveTorrents(ctx, item.Torrents)

	return nil
}

// List implements rms_library.ListsHandler.
func (s *Service) List(ctx context.Context, req *rms_library.ListsListRequest, resp *rms_library.ListsListResponse) error {
	items, err := s.Database.GetListItems(ctx, req.List, req.ContentType)
	if err != nil {
		logger.Errorf("List items failed: %s", err)
	}
	resp.Items = make([]*rms_library.ListItem, len(items))
	for i := range items {
		resp.Items[i] = &rms_library.ListItem{
			Id:          string(items[i].ID),
			Title:       items[i].Title,
			ContentType: items[i].ID.ContentType(),
			Size:        0, // TODO: calculate
		}
	}
	return nil
}

// Move implements rms_library.ListsHandler.
func (s *Service) Move(ctx context.Context, req *rms_library.ListsMoveRequest, resp *emptypb.Empty) error {
	item, err := s.Database.GetListItem(ctx, model.ID(req.Id))
	if err != nil {
		logger.Errorf("Cannot get '%s' item", req.Id, err)
		return err
	}
	if item == nil {
		return errors.New("not found")
	}

	if err = s.Database.MoveListItem(ctx, model.ID(req.Id), req.List); err != nil {
		logger.Errorf("Update item in db failed: %s", err)
		return err
	}

	s.Scheduler.Cancel(req.Id)

	return nil
}
