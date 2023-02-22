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

func (d Database) PutMovieInfo(ctx context.Context, id string, mov *rms_library.MovieInfo) error {
	record := model.Movie{
		ID:   id,
		Info: *mov,
	}

	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	opts := options.Replace().SetUpsert(true)
	filter := bson.D{{"_id", id}}

	_, err := d.cache.ReplaceOne(ctx, filter, &record, opts)
	return err
}

func (d Database) GetMovieInfo(ctx context.Context, id string) (*rms_library.MovieInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{{"_id", id}}
	result := d.cache.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, result.Err()
	}

	mov := model.Movie{}
	if err := result.Decode(&mov); err != nil {
		return nil, err
	}
	return &mov.Info, nil
}
