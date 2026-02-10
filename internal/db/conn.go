package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	cli   *mongo.Client
	db    *mongo.Database
	media *mongo.Collection
	cache *mongo.Collection
	meta  *mongo.Collection
}

const databaseTimeout = 40 * time.Second

const Version uint = 2

// Connect creates database connection
func Connect(uri string) (*Database, error) {
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

	db := &Database{
		cli:   cli,
		db:    lib,
		media: lib.Collection("media"),
		cache: lib.Collection("cache"),
		meta:  lib.Collection("metainfo"),
	}

	return db, nil
}
