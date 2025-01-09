package db

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
)

func (d Database) AddToWatchList(ctx context.Context, item *model.WatchListItem) error {
	_, err := d.watchlist.InsertOne(ctx, item)
	return err
}
