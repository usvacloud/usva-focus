package localredis

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/usvacloud/usva-focus/pkg/usva"
)

var Client *redis.Client
var once sync.Once

func Initialize() {
	once.Do(func() {
		Client = redis.NewClient(&redis.Options{})
	})
}

func Run(ctx context.Context) {
	RunPruner(ctx)
}

func Key(parts ...string) string {
	return strings.Join(
		append([]string{"usva", usva.Model, usva.Id}, parts...),
		":",
	)
}

func Zscore(ctx context.Context, zkey string, value string) (float64, error) {
	floatCmd := Client.ZScore(ctx, Key(zkey), value)
	if err := floatCmd.Err(); err != nil {
		return 0, err
	}

	return floatCmd.Result()
}
func Zlist(ctx context.Context, zkey string, count int64) []string {
	zRangeByScore := Client.ZRangeByScore(ctx, Key(zkey), &redis.ZRangeBy{
		Min: "-Inf", Max: "+Inf", Count: count,
	})
	if err := zRangeByScore.Err(); err != nil {
		log.Fatalln("Zlist", zkey, "Err", err)
	}
	members, err := zRangeByScore.Result()
	if err != nil {
		log.Fatalln("Zlist", zkey, "Result", err)
	}
	return members
}

func Zupdate(ctx context.Context, key string, value string) error {
	timestamp := time.Now().UTC().Unix()

	intCmd := Client.ZAdd(ctx, Key(key), &redis.Z{
		Score: float64(timestamp), Member: value,
	})
	return intCmd.Err()
}

func Zdelete(ctx context.Context, key string, value string) error {
	intCmd := Client.ZRem(ctx, Key(key), value)
	return intCmd.Err()
}

func BZrand(ctx context.Context, key string, delay time.Duration) string {
	for {
		select {
		case <-ctx.Done():
			return ""
		default:
			stringSliceCmd := Client.ZRandMember(ctx, Key(key), 1, false)
			if err := stringSliceCmd.Err(); err != nil {
				log.Fatalln("Zrand", "Cmd", err)
			}
			members, err := stringSliceCmd.Result()
			if err != nil {
				log.Fatalln("Zrand", "Result", err)
			}

			if len(members) > 0 {
				return members[0]
			}
			time.Sleep(delay)
		}
	}
}

func BZpopmin(ctx context.Context, key string) string {
	for {
		select {
		case <-ctx.Done():
			return ""
		default:
			zWithKeyCmd := Client.BZPopMin(ctx, time.Second, Key(key))
			if err := zWithKeyCmd.Err(); err != nil {
				continue
			}
			result, err := zWithKeyCmd.Result()
			if err != nil {
				log.Fatalln("localredis", "BZpopmin", "Result", err)
			}

			return result.Member.(string)
		}
	}
}
