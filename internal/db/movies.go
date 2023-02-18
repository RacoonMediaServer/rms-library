package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (d Database) GetDownloadedSeasons(ctx context.Context, id string) ([]uint, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	result := d.mov.FindOne(ctx, bson.D{{Key: "_id", Value: id}})

	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return []uint{}, nil
	}

	if result.Err() != nil {
		return []uint{}, result.Err()
	}

	mov := model.Movie{}
	if err := result.Decode(&mov); err != nil {
		return []uint{}, err
	}

	m := make(map[uint]struct{}, len(mov.Seasons))
	for i := range mov.Seasons {
		m[mov.Seasons[i].No] = struct{}{}
	}

	out := make([]uint, 0, len(m))
	for k, _ := range m {
		out = append(out, k)
	}

	return out, nil
}

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

func (d Database) UpdateMovieContent(mov *model.Movie) error {
	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()

	filter := bson.D{{"_id", mov.ID}}
	update := bson.D{{"$set", bson.D{{"files", mov.Files}, {"seasons", mov.Seasons}}}}
	_, err := d.mov.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
