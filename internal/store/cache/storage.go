package cache

import (
	"context"
	"social/internal/store"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	Users interface {
		Get(context.Context, int64) (*store.User, error)
		Set(context.Context, *store.User) error
	}
}

func NewRedisStore(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
	}

}