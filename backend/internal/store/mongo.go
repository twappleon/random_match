package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	Client *mongo.Client
	DB     *mongo.Database
}

func NewMongo(ctx context.Context, uri string, dbName string) (*Mongo, error) {
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(connectCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}
	return &Mongo{Client: client, DB: client.Database(dbName)}, nil
}

func (m *Mongo) Close(ctx context.Context) {
	_ = m.Client.Disconnect(ctx)
}
