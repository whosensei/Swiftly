package redis

import (
	"context"
	"fmt"
	"github/whosensei/shortenn/internal/utils"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
	Ctx    = context.Background()
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

	if err := Client.Ping(Ctx).Err(); err != nil {
		return nil,err
	}

	log.Println("Redis Connected")
	return Client, nil
}

func CheckRateLimit(key string, maxRequests int, window time.Duration) (bool, int, error) {
    rateLimitKey := fmt.Sprintf("ratelimit:%s", key)

    pipe := Client.TxPipeline()
    incrCmd := pipe.Incr(Ctx, rateLimitKey)
    pipe.Expire(Ctx, rateLimitKey, window)
    _, err := pipe.Exec(Ctx)

    if err != nil {
        return false, 0, err
    }

    count := int(incrCmd.Val())
    remaining := maxRequests - count

    return count <= maxRequests, remaining, nil
}
