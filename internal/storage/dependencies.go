package storage

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
)

type Database interface {
	GetListItem(ctx context.Context, id model.ID) (*model.ListItem, error)
}
