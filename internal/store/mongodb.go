package store

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoDBStore implements the Store interface with MongoDB as backend
type MongoDBStore struct {
	database    *mongo.Database
	collection  *mongo.Collection
	capacity    int
	insertCount atomic.Uint64
}

// NewMongoDBStore creates a new MongoDB-backend store
func NewMongoDBStore(uri, databaseName, collectionName string, capacity int) (*MongoDBStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB client: %w", err)
	}

	if err := client.Ping(connectCtx, readpref.Nearest()); err != nil {
		_ = client.Disconnect(connectCtx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(databaseName)
	collection := database.Collection(collectionName)

	indexName := fmt.Sprintf("%s_timestamp_idx", collectionName)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "timestamp", Value: -1}},
		Options: options.Index().SetName(indexName),
	}
	if _, err := collection.Indexes().CreateOne(connectCtx, indexModel); err != nil {
		_ = client.Disconnect(connectCtx)
		return nil, fmt.Errorf("failed to create index in MongoDB: %w", err)
	}

	return &MongoDBStore{
		database:   database,
		collection: collection,
		capacity:   capacity,
	}, nil
}

func (m *MongoDBStore) opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (m *MongoDBStore) Add(reqLog *model.RequestLog) error {
	ctx, cancel := m.opCtx()
	defer cancel()

	if _, err := m.collection.InsertOne(ctx, reqLog); err != nil {
		return fmt.Errorf("mongodb insert: %w", err)
	}
	if m.insertCount.Add(1)%cleanupEveryN == 0 {
		m.cleanup()
	}
	return nil
}

func (m *MongoDBStore) cleanup() {
	ctx, cancel := m.opCtx()
	defer cancel()

	// Select everything beyond the newest m.capacity documents by rank; a
	// separate count would go stale under concurrent inserts.
	findOptions := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(m.capacity)).
		SetProjection(bson.M{"_id": 1})

	cursor, err := m.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("govisual: failed to find oldest MongoDB logs: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var ids []string
	for cursor.Next(ctx) {
		var doc struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("govisual: failed to decode oldest MongoDB log: %v", err)
			continue
		}
		ids = append(ids, doc.ID)
	}
	if len(ids) == 0 {
		return
	}

	if _, err := m.collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}}); err != nil {
		log.Printf("govisual: failed to delete oldest MongoDB logs: %v", err)
	}
}

func (m *MongoDBStore) Get(id string) (*model.RequestLog, bool) {
	ctx, cancel := m.opCtx()
	defer cancel()

	var reqLog model.RequestLog
	if err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&reqLog); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, false
		}
		log.Printf("govisual: failed to get MongoDB log: %v", err)
		return nil, false
	}
	return &reqLog, true
}

func (m *MongoDBStore) GetAll() []*model.RequestLog {
	ctx, cancel := m.opCtx()
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := m.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		log.Printf("govisual: failed to query MongoDB: %v", err)
		return nil
	}
	defer cursor.Close(ctx)

	out := make([]*model.RequestLog, 0)
	for cursor.Next(ctx) {
		var reqLog model.RequestLog
		if err := cursor.Decode(&reqLog); err != nil {
			log.Printf("govisual: failed to decode MongoDB log: %v", err)
			continue
		}
		out = append(out, &reqLog)
	}
	return out
}

func (m *MongoDBStore) GetLatest(n int) []*model.RequestLog {
	ctx, cancel := m.opCtx()
	defer cancel()

	opts := options.Find().SetLimit(int64(n)).SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := m.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		log.Printf("govisual: failed to query MongoDB: %v", err)
		return nil
	}
	defer cursor.Close(ctx)

	out := make([]*model.RequestLog, 0)
	for cursor.Next(ctx) {
		var reqLog model.RequestLog
		if err := cursor.Decode(&reqLog); err != nil {
			log.Printf("govisual: failed to decode MongoDB log: %v", err)
			continue
		}
		out = append(out, &reqLog)
	}
	return out
}

func (m *MongoDBStore) Clear() error {
	ctx, cancel := m.opCtx()
	defer cancel()

	if _, err := m.collection.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("failed to clear MongoDB logs: %w", err)
	}
	return nil
}

func (m *MongoDBStore) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.database.Client().Disconnect(ctx)
}
