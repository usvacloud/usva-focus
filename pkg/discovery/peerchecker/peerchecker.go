package peerchecker

import (
	"context"
	"time"

	"github.com/usvacloud/usva-focus/pkg/protocol"
	"github.com/usvacloud/usva-focus/pkg/types"
)

func Run(ctx context.Context) {
	for {
		for _, peer := range types.Peers(ctx, 1) {
			protocol.Connect(ctx, peer.Url)
		}

		time.Sleep(1 * time.Second)
	}
}
