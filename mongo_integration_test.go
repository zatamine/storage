package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoPort           = "27017/tcp"
	mongoURI            = "mongodb://root:password@localhost"
	mongoTestDB         = "testDB"
	mongoTestCollection = "testCollection"
)

var dbClient *mongo.Client

var testMongoConfig mongoConfig

type item struct {
	Id   string `bson:"_id"`
	Name string `bson:"name"`
	Stat int64  `bson:"status"`
}

func (i item) ID() string {
	return i.Id
}
func (i item) Status() int64 {
	return i.Stat
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pull mongodb docker image for version 5.0
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "5.0",
		Env: []string{
			// username and password for mongodb superuser
			"MONGO_INITDB_ROOT_USERNAME=root",
			"MONGO_INITDB_ROOT_PASSWORD=password",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		dbClient, err = mongo.Connect(
			context.TODO(),
			options.Client().ApplyURI(
				fmt.Sprintf("%s:%s", mongoURI, resource.GetPort(mongoPort)),
			),
		)
		if err != nil {
			return err
		}
		return dbClient.Ping(context.TODO(), nil)
	})

	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	testMongoConfig = mongoConfig{
		mongoTestDB,
		fmt.Sprintf("%s:%s", mongoURI, resource.GetPort(mongoPort)),
	}

	// run tests
	code := m.Run()

	// When you're done, kill and remove the container
	if err = pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	// disconnect mongodb client
	if err = dbClient.Disconnect(context.TODO()); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func createItems(t *testing.T) []interface{} {
	t.Helper()
	db := dbClient.Database(mongoTestDB).Collection(mongoTestCollection)
	docs := []interface{}{
		bson.D{primitive.E{Key: "name", Value: "Bob"}},
		bson.D{primitive.E{Key: "name", Value: "Alice"}},
	}
	results, err := db.InsertMany(context.TODO(), docs)
	assert.NoError(t, err)
	return results.InsertedIDs
}
func tearDown(t *testing.T) {
	t.Helper()
	db := dbClient.Database(mongoTestDB).Collection(mongoTestCollection)
	err := db.Drop(context.TODO())
	require.NoError(t, err)
}
func TestCreate(t *testing.T) {
	defer tearDown(t)
	newItem := item{Name: "Bob"}
	ctx := context.Background()
	mongo, err1 := NewMongoDB[item](ctx, testMongoConfig, mongoTestCollection)
	require.NoError(t, err1)
	t.Run("Create item", func(t *testing.T) {
		err2 := mongo.Create(&newItem)
		assert.NoError(t, err2)
	})
}

func TestFindOne(t *testing.T) {
	defer tearDown(t)
	ctx := context.Background()
	ids := createItems(t)
	mongo, err1 := NewMongoDB[item](ctx, testMongoConfig, mongoTestCollection)
	require.NoError(t, err1)
	t.Run("Find one item", func(t *testing.T) {
		item, err2 := mongo.FindOne(ids[0].(primitive.ObjectID).Hex())
		assert.Equal(t, item.Name, "Bob")
		assert.NoError(t, err2)
	})
}

func TestUpdate(t *testing.T) {
	defer tearDown(t)
	ctx := context.Background()
	ids := createItems(t)
	mongo, err1 := NewMongoDB[item](ctx, testMongoConfig, mongoTestCollection)
	require.NoError(t, err1)
	t.Run("Update item", func(t *testing.T) {
		id := ids[0].(primitive.ObjectID).Hex()
		item := item{
			Id:   id,
			Name: "Bill",
			Stat: 200,
		}
		err2 := mongo.Update(item)
		assert.NoError(t, err2)
	})
}

func TestDelete(t *testing.T) {
	defer tearDown(t)
	ctx := context.Background()
	ids := createItems(t)
	mongo, err1 := NewMongoDB[item](ctx, testMongoConfig, mongoTestCollection)
	require.NoError(t, err1)
	t.Run("Delete item", func(t *testing.T) {
		err2 := mongo.Delete(ids[0].(primitive.ObjectID).Hex())
		assert.NoError(t, err2)
	})
}

func TestFindAll(t *testing.T) {
	defer tearDown(t)
	ctx := context.Background()
	ids := createItems(t)
	mongo, err1 := NewMongoDB[item](ctx, testMongoConfig, mongoTestCollection)
	require.NoError(t, err1)
	t.Run("Find all items", func(t *testing.T) {
		results, err2 := mongo.FindAll()
		assert.Equal(t, len(ids), len(results))
		assert.NoError(t, err2)
	})
}
