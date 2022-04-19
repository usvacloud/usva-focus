package types

import (
	"context"
	"errors"
	"log"
	"net/url"
	"time"

	"github.com/usvacloud/usva-focus/pkg/localredis"
)

type Peer struct {
	Url *url.URL
}

func NewPeer(urlString string) *Peer {
	u, err := url.Parse(urlString)
	if err != nil {
		log.Fatalln("url", err)
	}
	return &Peer{
		Url: u,
	}
}

func Peers(ctx context.Context, count int64) []Peer {
	peers := []Peer{}

	for _, urlString := range localredis.Zlist(ctx, "peers", count) {
		peers = append(peers, *NewPeer(urlString))
	}

	return peers
}

func GetRandomPeer(ctx context.Context) Peer {
	return *NewPeer(
		localredis.BZrand(ctx, "peers", 100*time.Millisecond),
	)
}

func (p Peer) Update(ctx context.Context) error {
	return localredis.Zupdate(ctx, "peers", p.Url.String())
}

func (p Peer) Delete(ctx context.Context) error {
	return localredis.Zdelete(ctx, "peers", p.Url.String())
}

func (p Peer) Since(ctx context.Context) (float64, error) {
	now := time.Now().UTC().Unix()

	score, err := localredis.Zscore(ctx, "peers", p.Url.String())
	if err != nil {
		return -1, errors.New("not found")
	}
	delta := float64(now) - score
	return delta, nil
}
