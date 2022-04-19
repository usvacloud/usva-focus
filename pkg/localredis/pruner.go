package localredis

import (
	"context"
	"log"
	"strconv"
	"time"
)

func RunPruner(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Println("pruner", "stopping ticker")
			ticker.Stop()
			log.Println("pruner", "ticker stopped")
			return
		case <-ticker.C:
			ago := time.Now().UTC().Unix() - int64(10)
			for _, zkey := range []string{"peers", "candidates", "selfs"} {
				intCmd := Client.ZRemRangeByScore(ctx, Key(zkey), "-inf", strconv.FormatInt(ago, 10))
				removed, _ := intCmd.Result()
				if removed > 0 {
					log.Println("pruner", "removed", removed, zkey)
				}
			}
		}
	}

}
