// Mongo storage implementation
package storage

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type mongoConfig struct {
	database      string
	connectionURI string
}

type mongoDB[T model] struct {
	ctx  context.Context
	coll mongo.Collection
	db   mongo.Database
}

func NewMongoDB[T model](ctx context.Context, config mongoConfig, collection string) (Storage[T], error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.connectionURI))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	db := *client.Database(config.database)
	return &mongoDB[T]{
		db:   db,
		coll: *db.Collection(collection),
	}, nil
}

func (m *mongoDB[T]) SetCollection(collection string) {
	m.coll = *m.db.Collection(collection)
}

func (m *mongoDB[T]) Collection() *mongo.Collection {
	return &m.coll
}

func (m *mongoDB[T]) FindOne(id string) (*T, error) {
	var item T
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.D{{Key: "_id", Value: objectID}}
	if err := m.coll.FindOne(m.ctx, filter).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (m *mongoDB[T]) FindAll() ([]T, error) {
	var item T
	items := make([]T, 0)
	filter := bson.D{{}}
	cursor, err := m.coll.Find(m.ctx, filter)
	if err != nil {
		defer cursor.Close(m.ctx)
		return items, err
	}

	for cursor.Next(m.ctx) {
		err := cursor.Decode(&item)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (m *mongoDB[T]) Create(item *T) error {
	_, err := m.coll.InsertOne(m.ctx, item)
	if err != nil {
		return err
	}
	return nil
}

// Update updates status field by ID
func (m *mongoDB[T]) Update(item T) error {
	id, err := primitive.ObjectIDFromHex(item.ID())
	if err != nil {
		return err
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	updater := bson.D{primitive.E{Key: "$set", Value: bson.D{
		primitive.E{Key: "status", Value: item.Status()},
	}}}
	result, err := m.coll.UpdateOne(m.ctx, filter, updater)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 || result.ModifiedCount == 0 {
		errMsg := fmt.Sprintf("Item not updated, matched count %d, modification count %d", result.MatchedCount, result.ModifiedCount)
		return errors.New(errMsg)
	}

	return nil
}
func (m *mongoDB[T]) Delete(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{primitive.E{Key: "_id", Value: objectID}}
	results, err := m.coll.DeleteOne(m.ctx, filter)
	if err != nil {
		return err
	}
	if results.DeletedCount == 0 {
		return errors.New("Cannot delete doccument, element not found.")
	}
	return nil
}
