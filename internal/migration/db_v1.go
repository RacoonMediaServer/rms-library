package migration

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type databaseV1 struct {
	cli       *mongo.Client
	db        *mongo.Database
	mov       *mongo.Collection
	watchlist *mongo.Collection
}

const databaseTimeout = 40 * time.Second

func connectV1(uri string) (*databaseV1, error) {
	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()

	cli, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect to db failed: %w", err)
	}

	if err = cli.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("connect to db failed: %w", err)
	}

	lib := cli.Database("library")

	db := &databaseV1{
		cli:       cli,
		db:        lib,
		mov:       lib.Collection("movies"),
		watchlist: lib.Collection("watchlist"),
	}

	return db, nil
}

func (d databaseV1) getWatchList(ctx context.Context) ([]*watchListItemV1, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"movieinfo.title", 1}})

	cur, err := d.watchlist.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var results []*watchListItemV1
	if err = cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (d databaseV1) getMovies(ctx context.Context) ([]*movieV1, error) {
	ctx, cancel := context.WithTimeout(ctx, databaseTimeout)
	defer cancel()

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"info.title", 1}})

	cur, err := d.mov.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var results []*movieV1
	if err = cur.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
