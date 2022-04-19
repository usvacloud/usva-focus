package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/usvacloud/usva-focus/pkg/commands"
	"github.com/usvacloud/usva-focus/pkg/daemon"
	"github.com/usvacloud/usva-focus/pkg/discovery"
	"github.com/usvacloud/usva-focus/pkg/localredis"
	"github.com/usvacloud/usva-focus/pkg/stats"
	"github.com/usvacloud/usva-focus/pkg/usva"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println("SIG", sig)
		cancel()

		wg.Wait()
		fmt.Println("all done")
	}()

	usva.Initialize()
	localredis.Initialize()
	usva.SpawnVoidFn(&wg, func() { localredis.Run(ctx) })
	usva.SpawnVoidFn(&wg, func() { stats.Run(ctx) })
	usva.SpawnVoidFn(&wg, func() { discovery.Run(ctx) })

	switch os.Args[1] {
	case "peer":
		log.Println("getting peer")
		commands.Peer(ctx)
	case "peers":
		log.Println("getting peers")
		commands.Peer(ctx)
		commands.Peers(ctx)

	case "daemon":
		usva.SpawnVoidFn(&wg, func() { daemon.Run(ctx) })
		wg.Wait()
	default:
		println("?")
	}

}
