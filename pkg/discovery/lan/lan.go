package lan

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/usvacloud/usva-focus/pkg/types"
)

func Resolve(ctx context.Context) {
	resolvCtx, cancel := context.WithTimeout(context.Background(), time.Second*1)

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			u := url.URL{
				Scheme: "http",
				Host:   entry.AddrIPv4[0].String() + ":" + fmt.Sprint(entry.Port),
			}
			log.Println("bonjour", u.String())
			types.NewCandidate(u.String()).Update(ctx)
		}
		cancel()
	}(entries)

	defer cancel()
	err = resolver.Browse(resolvCtx, "_usva._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	select {
	case <-ctx.Done():
		cancel()
	case <-resolvCtx.Done():
	}
}
