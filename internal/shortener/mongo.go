package shortener

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoStore is a MongoDB-backed link store. Links and click counts persist
// across restarts. All operations run under an explicit timeout so a slow or
// unreachable database never blocks a request indefinitely.
type MongoStore struct {
	client    *mongo.Client
	coll      *mongo.Collection
	opTimeout time.Duration
}

// NewMongoStore connects to MongoDB, verifies connectivity, and ensures the
// unique index on the short code. The caller owns the returned store's lifecycle
// and must call Close on shutdown.
func NewMongoStore(ctx context.Context, uri, dbName string, opTimeout time.Duration) (*MongoStore, error) {
	if opTimeout <= 0 {
		opTimeout = 5 * time.Second
	}

	// Bound connection + server selection so a bad URI fails fast rather than hanging.
	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(opTimeout).
		SetConnectTimeout(opTimeout)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	coll := client.Database(dbName).Collection("short_links")

	idxCtx, cancelIdx := context.WithTimeout(ctx, opTimeout)
	defer cancelIdx()
	_, err = coll.Indexes().CreateOne(idxCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "code", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ensure index: %w", err)
	}

	return &MongoStore{client: client, coll: coll, opTimeout: opTimeout}, nil
}

// Close disconnects the underlying client.
func (s *MongoStore) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

// Create validates rawURL and inserts it under a freshly generated unique code,
// retrying on the (rare) event of a code collision.
func (s *MongoStore) Create(rawURL string) (*Link, error) {
	normalized, verr := validateURL(rawURL)
	if verr != nil {
		return nil, verr
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.opTimeout)
	defer cancel()

	for attempts := 0; attempts < 10; attempts++ {
		code, err := randomCode()
		if err != nil {
			return nil, &Error{Code: "INTERNAL_ERROR", Message: "Could not allocate a short code"}
		}
		link := &Link{Code: code, URL: normalized, Clicks: 0, CreatedAt: time.Now().UTC()}

		_, err = s.coll.InsertOne(ctx, link)
		if err == nil {
			return link, nil
		}
		if mongo.IsDuplicateKeyError(err) {
			continue // code already taken; try another
		}
		return nil, &Error{Code: "INTERNAL_ERROR", Message: "Could not create short link"}
	}
	return nil, &Error{Code: "INTERNAL_ERROR", Message: "code space exhausted"}
}

// Resolve looks up a code and atomically increments its click counter, returning
// the updated link.
func (s *MongoStore) Resolve(code string) (*Link, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), s.opTimeout)
	defer cancel()

	var link Link
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err := s.coll.FindOneAndUpdate(
		ctx,
		bson.M{"code": code},
		bson.M{"$inc": bson.M{"clicks": 1}},
		opts,
	).Decode(&link)
	if err != nil {
		return nil, false
	}
	return &link, true
}

// Stats returns the link for a code without incrementing clicks.
func (s *MongoStore) Stats(code string) (*Link, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), s.opTimeout)
	defer cancel()

	var link Link
	if err := s.coll.FindOne(ctx, bson.M{"code": code}).Decode(&link); err != nil {
		return nil, false
	}
	return &link, true
}
