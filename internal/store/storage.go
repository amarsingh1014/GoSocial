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
	}

	Users interface {
		Create(context.Context, *User) error
	}

	Comments interface {
		GetByPostID(context.Context, int) ([]Comment, error)
	}
}

var (
	ErrNotFound = errors.New("resource not found")
)

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Posts : &PostsStore{db},
		Users : &UsersStore{db},
		Comments : &CommentsStore{db},
	}
}