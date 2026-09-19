package store

import (
	"context"

	"github.com/vignesh/livepoll/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PollRepo provides data access for polls.
type PollRepo struct {
	col *mongo.Collection
}

// Create inserts a new poll into the database.
func (r *PollRepo) Create(ctx context.Context, p *models.Poll) error {
	res, err := r.col.InsertOne(ctx, p)
	if isDuplicateKey(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}
	p.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByCode retrieves a poll by its share code.
func (r *PollRepo) FindByCode(ctx context.Context, code string) (*models.Poll, error) {
	var p models.Poll
	err := r.col.FindOne(ctx, bson.M{"code": code}).Decode(&p)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListByOwner returns all polls created by a specific user, newest first.
func (r *PollRepo) ListByOwner(ctx context.Context, ownerID primitive.ObjectID, limit int) ([]*models.Poll, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	
	cursor, err := r.col.Find(ctx, bson.M{"ownerId": ownerID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []*models.Poll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = make([]*models.Poll, 0)
	}
	return polls, nil
}

// IncrementTotals updates the durable vote counts for the specified options.
func (r *PollRepo) IncrementTotals(ctx context.Context, pollID primitive.ObjectID, optionIDs []string) error {
	inc := bson.M{"totalVotes": 1}
	for _, id := range optionIDs {
		inc["totals."+id] = 1
	}

	_, err := r.col.UpdateByID(ctx, pollID, bson.M{"$inc": inc})
	return err
}

// SetStatus updates a poll's status, but only if the owner matches.
func (r *PollRepo) SetStatus(ctx context.Context, pollID, ownerID primitive.ObjectID, status string) error {
	res, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": pollID, "ownerId": ownerID},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a poll, but only if the owner matches.
func (r *PollRepo) Delete(ctx context.Context, pollID, ownerID primitive.ObjectID) error {
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": pollID, "ownerId": ownerID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
