package redis

import (
	"context"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"os"
	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	ctx    = context.Background()
)

func InitRedis() (*redis.Client, error) {

	utils.LoadENV()

	redis_url := os.Getenv("REDIS_URL")
	if redis_url == "" {
		redis_url = "redis://localhost:6379"
	}
	
	opt, err := redis.ParseURL(redis_url)
	if err != nil {
		return nil,err
	} 

	Client = redis.NewClient(opt)

	if err := Client.Ping(ctx).Err(); err != nil {
		return nil,err
	}

	log.Println("Redis Connected")
	return Client, nil
}
