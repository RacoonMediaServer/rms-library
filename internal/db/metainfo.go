package db

import (
	"context"
	"errors"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const metaInfoKey = "metaInfo"

func (d Database) GetMetaInfo(ctx context.Context) (*model.MetaInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	result := d.meta.FindOne(ctx, bson.D{{Key: "_id", Value: metaInfoKey}})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return &model.MetaInfo{}, nil
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	mi := model.MetaInfo{}
	if err := result.Decode(&mi); err != nil {
		return nil, err
	}

	return &mi, nil
}

func (d Database) SetMetaInfo(ctx context.Context, mi model.MetaInfo) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	opts := options.Replace().SetUpsert(true)
	filter := bson.D{{"_id", metaInfoKey}}

	_, err := d.meta.ReplaceOne(ctx, filter, mi, opts)
	return err
}
