package stats

import (
	"context"
	"log"
	"time"

	"github.com/usvacloud/usva-focus/pkg/types"
)

func Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println(
				"peers", len(types.Peers(ctx, 100)),
				"candidates", len(types.Candidates(ctx, 100)),
			)
		}
	}
}
