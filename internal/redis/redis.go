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


func IncrementClicks(shortCode string) error {
    clickKey := fmt.Sprintf("clicks:%s", shortCode)
    return Client.Incr(Ctx, clickKey).Err()
}

func GetClickCount(shortCode string) (int64, error) {
    clickKey := fmt.Sprintf("clicks:%s", shortCode)
    count, err := Client.Get(Ctx, clickKey).Int64()
    if err == redis.Nil {
        return 0, nil
    }
    return count, err
}