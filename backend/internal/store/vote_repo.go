package store

import (
	"context"

	"github.com/vignesh/livepoll/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// VoteRepo provides data access for individual votes.
type VoteRepo struct {
	col *mongo.Collection
}

// Create inserts a new vote.
func (r *VoteRepo) Create(ctx context.Context, v *models.Vote) error {
	res, err := r.col.InsertOne(ctx, v)
	if isDuplicateKey(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}
	v.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// Exists checks if a vote for the given poll and voter already exists.
func (r *VoteRepo) Exists(ctx context.Context, pollID primitive.ObjectID, voterKey string) (bool, error) {
	err := r.col.FindOne(ctx, bson.M{"pollId": pollID, "voterKey": voterKey}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// CountByPoll reconstructs the tally from raw votes.
func (r *VoteRepo) CountByPoll(ctx context.Context, pollID primitive.ObjectID) (map[string]int64, int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"pollId": pollID}}},
		{{Key: "$facet", Value: bson.M{
			"totals": bson.A{
				bson.M{"$unwind": "$optionIds"},
				bson.M{"$group": bson.M{"_id": "$optionIds", "count": bson.M{"$sum": 1}}},
			},
			"totalVotes": bson.A{
				bson.M{"$count": "count"},
			},
		}}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, 0, nil
	}

	var result struct {
		Totals []struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		} `bson:"totals"`
		TotalVotes []struct {
			Count int64 `bson:"count"`
		} `bson:"totalVotes"`
	}

	if err := cursor.Decode(&result); err != nil {
		return nil, 0, err
	}

	counts := make(map[string]int64)
	for _, t := range result.Totals {
		counts[t.ID] = t.Count
	}

	var totalVotes int64
	if len(result.TotalVotes) > 0 {
		totalVotes = result.TotalVotes[0].Count
	}

	return counts, totalVotes, nil
}
