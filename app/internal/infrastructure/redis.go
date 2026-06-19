package infrastructure

import (
	"github.com/redis/go-redis/v9"
	"github.com/PKSonlem/testtask-secunda-api/internal/config"
)

func NewRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
