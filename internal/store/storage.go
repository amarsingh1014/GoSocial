package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
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
		Create(context.Context, *sql.Tx, *User) error
		GetById(context.Context, int) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string, invitationExpiry time.Duration) error
		Activate(ctx context.Context, token string) error
		Delete(ctx context.Context, id int64) error
		GetByEmail(ctx context.Context, email string) (*User, error)
	}

	Comments interface {
		GetByPostID(context.Context, int) ([]Comment, error)
		Create(context.Context, *Comment) error
	}

	Followers interface {
		Follow(ctx context.Context, followerId, userID int64) error
		Unfollow(ctx context.Context,followerId, userId int64) error
	}

	Roles interface {
		GetByName(context.Context, string) (*Role, error)
	}
}

var (
	ErrNotFound = errors.New("resource not found")
	ErrAlreadyFollowing = errors.New("already following")
	ErrNotFollowing = errors.New("not following")
	ErrDuplicateUsername = errors.New("duplicate username")
	ErrDuplicateEmail = errors.New("duplicate email")
)

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Posts : &PostsStore{db},
		Users : &UsersStore{db},
		Comments : &CommentsStore{db},
		Followers : &FollowersStore{db},
		Roles : &RolesStore{db},
	}
}

func withTx(db *sql.DB, ctx context.Context, f func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := f(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit() 
}