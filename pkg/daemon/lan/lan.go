package lan

import (
	"context"
	"log"

	"github.com/oleksandr/bonjour"
	"github.com/usvacloud/usva-focus/pkg/usva"
)

func Run(ctx context.Context) {
	usva.Wait(ctx, usva.DaemonStarted)

	server, err := bonjour.Register("usva focus "+usva.Id, "_usva._tcp", "", usva.PortInt, []string{"txtv=1", "app=usva-focus-" + usva.Id}, nil)
	if err != nil {
		panic(err)
	}

	log.Println("advertising as", usva.PortInt)
	<-ctx.Done()
	server.Shutdown()
}
