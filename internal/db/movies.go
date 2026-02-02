package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (d Database) GetOrCreateMovie(ctx context.Context, mov *model.Movie) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	result := d.mov.FindOne(ctx, bson.D{{Key: "_id", Value: mov.ID}})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		_, err := d.mov.InsertOne(ctx, mov)
		if err != nil {
			return fmt.Errorf("insert movie failed: %w", err)
		}

		return nil
	}

	if result.Err() != nil {
		return fmt.Errorf("fetch movie failed: %w", result.Err())
	}

	if err := result.Decode(mov); err != nil {
		return fmt.Errorf("decode movie record failed: %w", err)
	}

	return nil
}

func (d Database) AddMovie(ctx context.Context, mov *model.Movie) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	_, err := d.mov.InsertOne(ctx, mov)
	return err
}

func (d Database) UpdateMovieContent(ctx context.Context, mov *model.Movie) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{{"_id", mov.ID}}
	update := bson.D{{"$set", bson.D{{"torrents", mov.Torrents}, {"voice", mov.Voice}}}}
	_, err := d.mov.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (d Database) UpdateAvailableContent(ctx context.Context, mov *model.Movie) error {
	// ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	// defer cancel()

	// filter := bson.D{{"_id", mov.ID}}
	// update := bson.D{{"$set", bson.D{{"LastAvailableCheck", mov.LastAvailableCheck}, {"AvailableSeasons", mov.AvailableSeasons}}}}
	// _, err := d.mov.UpdateOne(ctx, filter, update)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (d Database) GetMovie(ctx context.Context, id string) (*model.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	result := d.mov.FindOne(ctx, bson.D{{Key: "_id", Value: id}})
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	if result.Err() != nil {
		return nil, result.Err()
	}

	mov := model.Movie{}
	if err := result.Decode(&mov); err != nil {
		return nil, err
	}

	return &mov, nil
}

func (d Database) SearchMovies(ctx context.Context, movieType *rms_library.MovieType) ([]*model.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{}
	if movieType != nil {
		filter = bson.D{{"info.type", int(*movieType)}}
	}
	opts := options.Find().SetSort(bson.D{{"info.title", 1}})

	cur, err := d.mov.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var results []*model.Movie
	if err = cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (d Database) DeleteMovie(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	_, err := d.mov.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	return err
}
