package daemon

import (
	"context"
	"log"
	"sync"

	"github.com/usvacloud/usva-focus/pkg/daemon/httpserver"
	"github.com/usvacloud/usva-focus/pkg/daemon/lan"
	"github.com/usvacloud/usva-focus/pkg/usva"
)

func Run(ctx context.Context) {
	var wg sync.WaitGroup

	usva.SpawnVoidFn(&wg, func() { httpserver.Run(ctx) })
	usva.SpawnVoidFn(&wg, func() { lan.Run(ctx) })

	usva.Wait(ctx, usva.DaemonStarted)

	log.Println("usva focus", usva.Id, usva.Port)

	<-ctx.Done()

	log.Println("daemon shut down start")
	wg.Wait()
	log.Println("daemon shut down done")
}
