package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Database struct {
	cli *mongo.Client
	db  *mongo.Database
	mov *mongo.Collection
}

const databaseTimeout = 40 * time.Second

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

	db := &Database{
		cli: cli,
		db:  cli.Database("library"),
		mov: cli.Database("library").Collection("movies"),
	}

	return db, nil
}
