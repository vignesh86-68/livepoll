// Package store is the MongoDB persistence layer.
//
// Nothing in this package knows what HTTP is. Repositories take and return
// domain types from the models package and report failures as plain errors, so
// the handlers above can be tested without a database and the database code can
// be changed without touching the handlers.
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ErrNotFound is returned when a lookup matched no document. Callers check for
// this rather than comparing against a driver-specific sentinel, which keeps
// mongo-driver out of the layers above.
var ErrNotFound = errors.New("not found")

// ErrDuplicate is returned when a write violated a unique index.
var ErrDuplicate = errors.New("duplicate")

// Store owns the MongoDB client and exposes one repository per collection.
type Store struct {
	client *mongo.Client
	db     *mongo.Database

	Users *UserRepo
	Polls *PollRepo
	Votes *VoteRepo
}

// Connect dials MongoDB and verifies the connection before returning.
//
// The driver connects lazily, so without an explicit Ping a bad URI or a
// missing network allowlist entry would not surface until the first real
// request — most likely as a confusing timeout in front of a user. Failing at
// startup instead makes a misconfiguration obvious in the deploy logs.
func Connect(ctx context.Context, uri, dbName string) (*Store, error) {
	opts := options.Client().
		ApplyURI(uri).
		SetAppName("livepoll").
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetMaxPoolSize(50).
		SetRetryWrites(true)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("store: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("store: ping: %w", err)
	}

	db := client.Database(dbName)
	return &Store{
		client: client,
		db:     db,
		Users:  &UserRepo{col: db.Collection("users")},
		Polls:  &PollRepo{col: db.Collection("polls")},
		Votes:  &VoteRepo{col: db.Collection("votes")},
	}, nil
}

// EnsureIndexes creates every index the application depends on.
//
// Two of these are not performance tuning but correctness guarantees, and the
// distinction is worth being clear about. The unique index on users.email and
// the unique index on (votes.pollId, votes.voterKey) are what actually prevent
// duplicate accounts and double voting. Checking "does this already exist?" in
// application code before inserting cannot do that job: two concurrent requests
// can both read "no" before either writes. Only the database, at the point of
// the write, can settle the race.
//
// Creating an index that already exists with the same specification is a no-op,
// so this is safe to run on every boot.
func (s *Store) EnsureIndexes(ctx context.Context) error {
	indexes := []struct {
		collection string
		model      mongo.IndexModel
	}{
		{"users", mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_email"),
		}},
		{"polls", mongo.IndexModel{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_code"),
		}},
		{"polls", mongo.IndexModel{
			// Serves the "my polls" listing: filter by owner, newest first.
			Keys:    bson.D{{Key: "ownerId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("owner_recent"),
		}},
		{"votes", mongo.IndexModel{
			Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterKey", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uniq_poll_voter"),
		}},
		{"votes", mongo.IndexModel{
			// Serves tally rebuilds and any future "recent votes" view.
			Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("poll_recent"),
		}},
	}

	for _, ix := range indexes {
		if _, err := s.db.Collection(ix.collection).Indexes().CreateOne(ctx, ix.model); err != nil {
			return fmt.Errorf("store: create index on %s: %w", ix.collection, err)
		}
	}
	return nil
}

// Ping reports whether MongoDB is reachable. Used by the health endpoint so the
// platform's health check fails when a dependency is down, rather than the app
// happily serving requests it cannot fulfil.
func (s *Store) Ping(ctx context.Context) error {
	return s.client.Ping(ctx, readpref.Primary())
}

// Close disconnects the client.
func (s *Store) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

// isDuplicateKey reports whether err came from violating a unique index.
func isDuplicateKey(err error) bool {
	return err != nil && mongo.IsDuplicateKeyError(err)
}
