package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

func InitRedis() error {

	redis_url := os.Getenv("REDIS_URL")
	if redis_url == "" {
		redis_url = "redis://localhost:6379"
	}
	
	opt, err := redis.ParseURL(redis_url)
	if err != nil {
		return err
	} 

	Client = redis.NewClient(opt)

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}
