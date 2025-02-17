package db

import (
	"context"
	"errors"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d Database) AddToWatchList(ctx context.Context, item *model.WatchListItem) error {
	_, err := d.watchlist.InsertOne(ctx, item)
	return err
}

func (d Database) GetWatchList(ctx context.Context, movieType *rms_library.MovieType) ([]*model.WatchListItem, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{}
	if movieType != nil {
		filter = bson.D{{"movieinfo.type", int(*movieType)}}
	}
	opts := options.Find().SetSort(bson.D{{"movieinfo.title", 1}})

	cur, err := d.watchlist.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var results []*model.WatchListItem
	if err = cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (d Database) GetWatchListItem(ctx context.Context, id string) (*model.WatchListItem, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	result := d.watchlist.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	mov := model.WatchListItem{}
	if err := result.Decode(&mov); err != nil {
		return nil, err
	}

	return &mov, nil
}
