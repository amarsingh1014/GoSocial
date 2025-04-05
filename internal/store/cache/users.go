package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"social/internal/store"
	"time"

	"github.com/go-redis/redis/v8"
)

type UserStore struct {
	rdb *redis.Client
}

var UserExpiry = time.Minute * 5

func (s *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)

	data, err := s.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err
	}

	var user store.User
	if data != "" {
		err := json.Unmarshal([]byte(data), &user)
		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (s *UserStore) Set(ctx context.Context, user *store.User) error {
	cacheKey := fmt.Sprintf("user:%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return s.rdb.SetEX(ctx, cacheKey, data, UserExpiry).Err()
}