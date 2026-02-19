package db

import (
	"context"
	"errors"

	"github.com/RacoonMediaServer/rms-library/v3/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d Database) GetListItems(ctx context.Context, list *rms_library.List, contentType *rms_library.ContentType, sort *rms_library.Sort, p *rms_library.Pagination) (results []*model.ListItem, err error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{}
	if list != nil {
		filter = append(filter, bson.E{Key: "list", Value: int(*list)})
	}
	if contentType != nil {
		filter = append(filter, bson.E{Key: "contenttype", Value: int(*contentType)})
	}

	opts := options.Find().SetSort(getSort(sort))
	if p != nil {
		opts.SetSkip(int64(p.Offset)).SetLimit(int64(p.Count))
	}

	var cur *mongo.Cursor
	cur, err = d.media.Find(ctx, filter, opts)
	if err != nil {
		return
	}

	if err = cur.All(ctx, &results); err != nil {
		return
	}

	return
}

func (d Database) MoveListItem(ctx context.Context, id model.ID, newList rms_library.List) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{{"_id", id.String()}}
	update := bson.D{{"$set", bson.D{{"list", newList}}}}
	_, err := d.media.UpdateOne(ctx, filter, update)
	return err
}

func (d Database) DeleteListItem(ctx context.Context, id model.ID) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	_, err := d.media.DeleteOne(ctx, bson.D{{Key: "_id", Value: id.String()}})
	return err
}

func (d Database) GetListItem(ctx context.Context, id model.ID) (*model.ListItem, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	result := d.media.FindOne(ctx, bson.D{{Key: "_id", Value: id.String()}})
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
	_, err := d.media.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
