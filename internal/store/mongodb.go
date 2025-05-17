package store

import (
	"context"
	"fmt"
	"log"

	"github.com/doganarif/govisual/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoDBStore implements the Store interface with MongoDB as backend
type MongoDBStore struct {
	database   *mongo.Database
	collection *mongo.Collection
	capacity   int
	ctx        context.Context
}

// NewMongoDBStore creates a new MongoDB-backend store
func NewMongoDBStore(uri, databaseName, collectionName string, capacity int) (*MongoDBStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	ctx := context.Background()
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to get MongoDB client: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, readpref.Nearest()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(databaseName)
	collection := database.Collection(collectionName)

	// Create index on timestamp for faster retrieval
	indexName := fmt.Sprintf("%s_timestamp_idx", collectionName)
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"Timestamp": -1},
		Options: options.Index().SetName(indexName),
	}

	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create index in MongoDB: %w", err)
	}
	return &MongoDBStore{
		database:   database,
		collection: collection,
		capacity:   capacity,
		ctx:        ctx,
	}, nil
}

// Add adds a new request log to the store
func (m *MongoDBStore) Add(reqLog *model.RequestLog) {
	// Store log in MongoDB
	if _, err := m.collection.InsertOne(m.ctx, reqLog); err != nil {
		log.Printf("Failed to store log in MongoDB: %v", err)
		return
	}
	m.cleanup()
}

// cleanup removes old logs to maintain the capacity limit
func (m *MongoDBStore) cleanup() {
	count, err := m.collection.CountDocuments(m.ctx, bson.M{})
	if err != nil {
		log.Printf("Failed to get the log count in MongoDB: %v", err)
		return
	}

	if count <= int64(m.capacity) {
		return
	}
	// Find the oldest logs that exceed capacity
	findOptions := options.Find().
		SetSort(bson.D{{Key: "Timestamp", Value: 1}}).
		SetLimit(count - int64(m.capacity))

	cursor, err := m.collection.Find(m.ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("Failed to find oldest logs in MongoDB: %v", err)
		return
	}
	defer cursor.Close(m.ctx)

	var oldestLogs []model.RequestLog
	for cursor.Next(m.ctx) {
		var reqLog model.RequestLog
		if err := cursor.Decode(&reqLog); err != nil {
			log.Printf("Failed to decode oldest log in MongoDB: %v", err)
			continue
		}
		oldestLogs = append(oldestLogs, reqLog)
	}

	if len(oldestLogs) == 0 {
		return
	}

	// Extract IDs of logs to delete
	var ids []string
	for _, log := range oldestLogs {
		ids = append(ids, log.ID)
	}

	// Delete the oldest logs
	if _, err := m.collection.DeleteMany(m.ctx, bson.M{"_id": bson.M{"$in": ids}}); err != nil {
		log.Printf("Failed to delete oldest logs in MongoDB: %v", err)
		return
	}
}

// Get retrieves a specific request log by its ID
func (m *MongoDBStore) Get(id string) (*model.RequestLog, bool) {
	var reqLog model.RequestLog
	if err := m.collection.FindOne(m.ctx, bson.M{"_id": id}).Decode(&reqLog); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, false
		}
		log.Printf("Failed to get request log from MongoDB: %v", err)
		return nil, false
	}
	return &reqLog, true
}

// GetAll returns all stored request logs
func (m *MongoDBStore) GetAll() []*model.RequestLog {
	opts := options.Find().SetSort(bson.M{"Timestamp": -1})
	cursor, err := m.collection.Find(m.ctx, bson.M{}, opts)
	if err != nil {
		if err == mongo.ErrClientDisconnected {
			return nil
		}
		log.Printf("Failed to get cursor from MongoDB: %v", err)
		return nil
	}
	defer cursor.Close(m.ctx)
	reqsLog := make([]*model.RequestLog, 0)
	for cursor.Next(m.ctx) {
		var reqLog model.RequestLog
		if err := cursor.Decode(&reqLog); err != nil {
			log.Printf("Failed to decode request log from MongoDB: %v", err)
			continue
		}
		reqsLog = append(reqsLog, &reqLog)
	}
	return reqsLog
}

// GetLatest returns the n most recent request logs
func (m *MongoDBStore) GetLatest(n int) []*model.RequestLog {
	// Get the n newest log IDs
	opts := options.Find().SetLimit(int64(n)).SetSort(bson.M{"timestamp": -1})
	cursor, err := m.collection.Find(m.ctx, bson.M{}, opts)
	if err != nil {
		if err == mongo.ErrClientDisconnected {
			return nil
		}
		log.Printf("Failed to get cursor from MongoDB: %v", err)
		return nil
	}
	defer cursor.Close(m.ctx)
	reqsLog := make([]*model.RequestLog, 0)
	for cursor.Next(m.ctx) {
		var reqLog model.RequestLog
		if err := cursor.Decode(&reqLog); err != nil {
			log.Printf("Failed to decode request log from MongoDB: %v", err)
			continue
		}
		reqsLog = append(reqsLog, &reqLog)
	}

	return reqsLog
}

// Clear removes all logs from the store
func (m *MongoDBStore) Clear() error {
	_, err := m.collection.DeleteMany(m.ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to clear logs in MongoDB: %w", err)
	}
	return nil
}

// Close closes the database connection
func (m *MongoDBStore) Close() error {
	return m.database.Client().Disconnect(m.ctx)
}
