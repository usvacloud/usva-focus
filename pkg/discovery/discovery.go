package discovery

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/usvacloud/usva-focus/pkg/discovery/candidatepromoter"
	"github.com/usvacloud/usva-focus/pkg/discovery/lan"
	"github.com/usvacloud/usva-focus/pkg/discovery/peerchecker"
	"github.com/usvacloud/usva-focus/pkg/types"
)

func Run(ctx context.Context) {
	go candidatepromoter.Run(ctx)
	go lan.Resolve(ctx)
	go fromSeeds(ctx)
	go peerchecker.Run(ctx)

	lanTicker := time.NewTicker(5 * time.Second)
	seedTicker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			log.Println("discovery", "done")
			return
		case <-lanTicker.C:
			lan.Resolve(ctx)
		case <-seedTicker.C:
			fromSeeds(ctx)
		}
	}
}

func fromSeeds(ctx context.Context) {
	if value := os.Getenv("USVA_SEEDS"); value != "" {
		types.NewCandidate(
			value,
		).Update(ctx)
	}
}
