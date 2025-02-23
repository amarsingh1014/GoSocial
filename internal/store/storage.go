package store

import (
	"context"
	"database/sql"
	"errors"
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetById(context.Context, int) (*Post, error)
		Delete(context.Context, int) error
		Update(context.Context, *Post) error
		GetUserFeed(context.Context, int64, PaginatedFieldQuery) ([]PostWithMetadata, error)
	}

	Users interface {
		Create(context.Context, *User) error
		GetById(context.Context, int) (*User, error)
	}

	Comments interface {
		GetByPostID(context.Context, int) ([]Comment, error)
		Create(context.Context, *Comment) error
	}

	Followers interface {
		Follow(ctx context.Context, followerId, userID int64) error
		Unfollow(ctx context.Context,followerId, userId int64) error
	}
}

var (
	ErrNotFound = errors.New("resource not found")
	ErrAlreadyFollowing = errors.New("already following")
	ErrNotFollowing = errors.New("not following")
)

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Posts : &PostsStore{db},
		Users : &UsersStore{db},
		Comments : &CommentsStore{db},
		Followers : &FollowersStore{db},
	}
}