package rediscache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PKSonlem/testtask-secunda-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

const ttl = 5 * time.Minute

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) GetTeamTasks(ctx context.Context, key string) ([]*domain.Task, bool, error) {
	data, err := c.client.Get(ctx, key).Bytes()

	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var tasks []*domain.Task

	return tasks, true, json.Unmarshal(data, &tasks)
}

func (c *Cache) SetTeamTasks(ctx context.Context, key string, tasks []*domain.Task) error {
	data, err := json.Marshal(tasks)

	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) Invalidate(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
