package candidatepromoter

import (
	"context"
	"log"

	"github.com/usvacloud/usva-focus/pkg/protocol"
	"github.com/usvacloud/usva-focus/pkg/types"
)

func Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("http discovery done")
			return
		default:
			candidate := types.PopCandidate(ctx)

			protocol.Connect(ctx, candidate.Url)
		}

	}
}
