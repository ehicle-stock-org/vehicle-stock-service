package mongo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/vehicle-stock-service/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Client *mongo.Client

// For testability
var MongoConnectFunc = func(ctx context.Context, uri string) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI(uri))
}
var MongoPingFunc = func(client *mongo.Client, ctx context.Context) error {
	return client.Ping(ctx, readpref.Primary())
}

// Add function variable for InsertOne
var MongoInsertOneFunc = func(coll *mongo.Collection, ctx context.Context, data interface{}) (interface{}, error) {
	return coll.InsertOne(ctx, data)
}

// ConnectMongo connects to MongoDB Atlas, allows mocking for tests
func ConnectMongo(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := MongoConnectFunc(ctx, uri)
	if err != nil {
		return nil, err
	}

	if err := MongoPingFunc(client, ctx); err != nil {
		return nil, err
	}

	Client = client
	fmt.Println("Connected to MongoDB!")
	return client, nil
}

// InsertData inserts a record into the collection (generic)
func InsertData(database, collection string, data interface{}) error {
	if Client == nil {
		return fmt.Errorf("Mongo client is not initialized")
	}
	coll := Client.Database(database).Collection(collection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := MongoInsertOneFunc(coll, ctx, data)
	if err != nil {
		log.Println("Error inserting data:", err)
		return err
	}
	log.Println("Inserted data successfully:", data)
	return nil
}

// Define interfaces for testability
type Collection interface {
	FindOne(ctx context.Context, filter interface{}) SingleResult
}

type SingleResult interface {
	Decode(v interface{}) error
}

// Adapter for mongo.Collection
type mongoCollectionAdapter struct {
	coll *mongo.Collection
}

func (m *mongoCollectionAdapter) FindOne(ctx context.Context, filter interface{}) SingleResult {
	return &mongoSingleResultAdapter{res: m.coll.FindOne(ctx, filter)}
}

// Adapter for mongo.SingleResult
type mongoSingleResultAdapter struct {
	res *mongo.SingleResult
}

func (m *mongoSingleResultAdapter) Decode(v interface{}) error {
	return m.res.Decode(v)
}

// Update FindStockByTickerAndDate to use adapters
var FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
	if Client == nil {
		return nil, fmt.Errorf("Mongo client is not initialized")
	}
	coll := &mongoCollectionAdapter{coll: Client.Database(database).Collection(collection)}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := map[string]interface{}{
		"ticker": ticker,
		"time":   date,
	}
	var result models.StockData
	singleResult := coll.FindOne(ctx, filter)
	if err := singleResult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Exported for testability in other packages
var InsertDataFunc = InsertData
