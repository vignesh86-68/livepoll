package store

import (
	"context"

	"github.com/vignesh/livepoll/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepo provides data access for users.
type UserRepo struct {
	col *mongo.Collection
}

// Create inserts a new user into the database.
func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	res, err := r.col.InsertOne(ctx, u)
	if isDuplicateKey(err) {
		return ErrDuplicate
	}
	if err != nil {
		return err
	}
	u.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// FindByEmail retrieves a user by their email address.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}
