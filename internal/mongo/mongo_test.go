package mongo

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/vehicle-stock-service/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mocks for error branch testing
type mockSingleResult struct {
	decodeErr error
}

func (r *mockSingleResult) Decode(v interface{}) error {
	return r.decodeErr
}

type mockCollection struct {
	findErr, decodeErr error
}

func (c *mockCollection) FindOne(ctx context.Context, filter interface{}) SingleResult {
	if c.findErr != nil {
		return &mockSingleResult{decodeErr: c.findErr}
	}
	return &mockSingleResult{decodeErr: c.decodeErr}
}

const findErrMsg = "find error"
const decodeErrMsg = "decode error"

const testDate = "2025-08-24"

// --- ConnectMongo tests ---
func TestConnectMongoSuccess(t *testing.T) {
	origConnect := MongoConnectFunc
	origPing := MongoPingFunc
	defer func() { MongoConnectFunc = origConnect; MongoPingFunc = origPing }()
	MongoConnectFunc = func(ctx context.Context, uri string) (*mongo.Client, error) {
		return &mongo.Client{}, nil
	}
	MongoPingFunc = func(client *mongo.Client, ctx context.Context) error {
		return nil
	}
	_, err := ConnectMongo("mongodb://localhost:27017")
	assert.NoError(t, err)
}

func TestConnectMongoInvalidURI(t *testing.T) {
	origConnect := MongoConnectFunc
	defer func() { MongoConnectFunc = origConnect }()
	MongoConnectFunc = func(ctx context.Context, uri string) (*mongo.Client, error) {
		return nil, errors.New("invalid uri")
	}
	_, err := ConnectMongo("")
	assert.Error(t, err)
}

func TestConnectMongoPingError(t *testing.T) {
	origConnect := MongoConnectFunc
	origPing := MongoPingFunc
	defer func() { MongoConnectFunc = origConnect; MongoPingFunc = origPing }()
	MongoConnectFunc = func(ctx context.Context, uri string) (*mongo.Client, error) {
		return &mongo.Client{}, nil
	}
	MongoPingFunc = func(client *mongo.Client, ctx context.Context) error {
		return errors.New("ping error")
	}
	_, err := ConnectMongo("mongodb://localhost:27017")
	assert.Error(t, err)
}

// --- InsertData tests ---
func TestInsertDataSuccess(t *testing.T) {
	// Mock Client and collection
	Client = &mongo.Client{}
	// This will error unless a real MongoDB is running, so we expect error
	err := InsertData("db", "coll", map[string]interface{}{"foo": "bar"})
	assert.Error(t, err)
}

func TestInsertDataInsertError(t *testing.T) {
	// Simulate error by setting Client to nil
	Client = nil
	err := InsertData("db", "coll", map[string]interface{}{"foo": "bar"})
	assert.Error(t, err)
}

// --- InsertData full branch coverage ---
func TestInsertDataBranches(t *testing.T) {
	origClient := Client
	origInsert := MongoInsertOneFunc
	defer func() { Client = origClient; MongoInsertOneFunc = origInsert }()

	t.Run("nil client", func(t *testing.T) {
		Client = nil
		err := InsertData("db", "coll", map[string]interface{}{"foo": "bar"})
		assert.Error(t, err)
	})

	t.Run("insert error", func(t *testing.T) {
		const findErrMsg = "find error"
		const decodeErrMsg = "decode error"

		type mockSingleResult struct {
			decodeErr error
		}
		Client = &mongo.Client{} // Just to pass nil check
		MongoInsertOneFunc = func(coll *mongo.Collection, ctx context.Context, data interface{}) (interface{}, error) {
			return nil, errors.New("insert fail")
		}
		err := InsertData("db", "coll", map[string]interface{}{"foo": "bar"})
		assert.Error(t, err)
	})

	t.Run("success path", func(t *testing.T) {
		Client = &mongo.Client{}
		MongoInsertOneFunc = func(coll *mongo.Collection, ctx context.Context, data interface{}) (interface{}, error) {
			return struct{}{}, nil
		}
		err := InsertData("db", "coll", map[string]interface{}{"foo": "bar"})
		assert.NoError(t, err)
	})
}

// --- FindStockByTickerAndDate tests ---
func TestFindStockByTickerAndDateSuccess(t *testing.T) {
	// Mock FindStockByTickerAndDate to return success
	orig := FindStockByTickerAndDate
	FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
		return &models.StockData{Ticker: ticker, Bid: 100.0, Ask: 101.0, Time: date}, nil
	}
	defer func() { FindStockByTickerAndDate = orig }()
	result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
	assert.NoError(t, err)
	assert.Equal(t, "AAPL", result.Ticker)
	assert.Equal(t, 100.0, result.Bid)
	assert.Equal(t, 101.0, result.Ask)
	assert.Equal(t, testDate, result.Time)
}

func TestFindStockByTickerAndDateFindError(t *testing.T) {
	orig := FindStockByTickerAndDate
	FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
		return nil, errors.New("find error")
	}
	defer func() { FindStockByTickerAndDate = orig }()
	result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFindStockByTickerAndDateDecodeError(t *testing.T) {
	orig := FindStockByTickerAndDate
	FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
		return nil, errors.New("decode error")
	}
	defer func() { FindStockByTickerAndDate = orig }()
	result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- FindStockByTickerAndDate full branch coverage ---
func TestFindStockByTickerAndDateBranches(t *testing.T) {
	orig := FindStockByTickerAndDate
	defer func() { FindStockByTickerAndDate = orig }()

	t.Run("nil client", func(t *testing.T) {
		Client = nil
		result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("find error", func(t *testing.T) {
		FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
			return nil, errors.New("find error")
		}
		result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("decode error", func(t *testing.T) {
		FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
			return nil, errors.New("decode error")
		}
		result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("success path", func(t *testing.T) {
		FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
			return &models.StockData{Ticker: ticker, Bid: 100.0, Ask: 101.0, Time: date}, nil
		}
		result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "AAPL", result.Ticker)
		assert.Equal(t, testDate, result.Time)
	})
}

// --- Adapter-based error branch coverage for FindStockByTickerAndDate ---
func TestFindStockByTickerAndDate_AdapterErrorBranches(t *testing.T) {
	const findErrMsg = "find error"

	FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
		coll := &mockCollection{findErr: errors.New(findErrMsg), decodeErr: nil}
		ctx := context.Background()
		var result models.StockData
		singleResult := coll.FindOne(ctx, map[string]interface{}{})
		if err := singleResult.Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), findErrMsg)
	assert.Nil(t, result)

	FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
		coll := &mockCollection{findErr: nil, decodeErr: errors.New(decodeErrMsg)}
		ctx := context.Background()
		var result models.StockData
		singleResult := coll.FindOne(ctx, map[string]interface{}{})
		if err := singleResult.Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	result, err = FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), decodeErrMsg)
	assert.Nil(t, result)
}

func TestFindStockByTickerAndDate_ErrorBranches(t *testing.T) {

	const findErrMsg = "find error"
	const decodeErrMsg = "decode error"

	FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
		coll := &mockCollection{findErr: errors.New(findErrMsg), decodeErr: nil}
		ctx := context.Background()
		var result models.StockData
		singleResult := coll.FindOne(ctx, map[string]interface{}{})
		if err := singleResult.Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	result, err := FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), findErrMsg)
	assert.Nil(t, result)

	FindStockByTickerAndDate = func(database, collection, ticker, date string) (*models.StockData, error) {
		coll := &mockCollection{findErr: nil, decodeErr: errors.New(decodeErrMsg)}
		ctx := context.Background()
		var result models.StockData
		singleResult := coll.FindOne(ctx, map[string]interface{}{})
		if err := singleResult.Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil
	}
	result, err = FindStockByTickerAndDate("db", "coll", "AAPL", testDate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), decodeErrMsg)
	assert.Nil(t, result)
}
