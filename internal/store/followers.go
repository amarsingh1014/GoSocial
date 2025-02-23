package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Follower struct {
	UserID int64 `json:"user_id"`
	FollowerID int64 `json:"follower_id"`
	CreatedAt string `json:"created_at"`
}

type FollowersStore struct {
	db *sql.DB
}

func (s *FollowersStore) Follow(ctx context.Context, followerId, userID int64) error {
	// Follow a user
	query := `INSERT INTO followers (user_id, follower_id) VALUES ($1, $2)`

	_, err := s.db.ExecContext(ctx, query, userID, followerId)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrAlreadyFollowing
		}
	}

	return nil
}

func (s *FollowersStore) Unfollow(ctx context.Context, followerId, userId int64) error {
	// Unfollow a user
	query := `
		DELETE FROM followers 
		WHERE user_id = $1 
		AND follower_id = $2`

	_, err := s.db.ExecContext(ctx, query, userId, followerId)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrNotFollowing
		default:
			return err
		}
	}

	return nil
}
