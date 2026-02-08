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

func (d Database) GetListItems(ctx context.Context, list rms_library.List, contentType *rms_library.ContentType) (results []*model.ListItem, err error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{}
	filter = bson.D{{"list", int(list)}}

	opts := options.Find().SetSort(bson.D{{"title", 1}})

	var cur *mongo.Cursor
	cur, err = d.mov.Find(ctx, filter, opts)
	if err != nil {
		return
	}

	if err = cur.All(ctx, &results); err != nil {
		return
	}

	// TODO: different collections
	return
}

func (d Database) MoveListItem(ctx context.Context, id model.ID, newList rms_library.List) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	// TODO: different collections

	filter := bson.D{{"_id", id.String()}}
	update := bson.D{{"$set", bson.D{{"list", newList}}}}
	_, err := d.mov.UpdateOne(ctx, filter, update)
	return err
}

func (d Database) DeleteListItem(ctx context.Context, id model.ID) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	// TODO: different collections
	_, err := d.mov.DeleteOne(ctx, bson.D{{Key: "_id", Value: id.String()}})
	return err
}

func (d Database) GetListItem(ctx context.Context, id model.ID) (*model.ListItem, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	// TODO: different collections

	result := d.mov.FindOne(ctx, bson.D{{Key: "_id", Value: id.String()}})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	item := model.ListItem{}
	if err := result.Decode(&item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (d Database) UpdateContent(ctx context.Context, id model.ID, torrents []model.TorrentRecord) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{{"_id", id.String()}}
	update := bson.D{{"$set", bson.D{{"torrents", torrents}}}}
	_, err := d.mov.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// TODO: different collections

	return nil
}
