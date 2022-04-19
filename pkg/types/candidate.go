package types

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/usvacloud/usva-focus/pkg/localredis"
)

type Candidate struct {
	Url *url.URL
}

func NewCandidate(urlString string) *Candidate {
	u, err := url.Parse(urlString)
	if err != nil {
		log.Fatalln("url", err)
	}
	return &Candidate{
		Url: u,
	}
}
func Candidates(ctx context.Context, count int64) []*Candidate {
	candidates := []*Candidate{}

	for _, urlString := range localredis.Zlist(ctx, "candidates", count) {
		candidates = append(candidates, NewCandidate(urlString))
	}
	return candidates
}

func PopCandidate(ctx context.Context) *Candidate {
	return NewCandidate(
		localredis.BZpopmin(ctx, "candidates"),
	)
}

func (c Candidate) Update(ctx context.Context) error {
	maybePeer := NewPeer(c.Url.String())
	since, err := maybePeer.Since(ctx)
	if err == nil {
		return errors.New("candidate already a peer for " + fmt.Sprint(since) + "s")
	}

	return localredis.Zupdate(ctx, "candidates", c.Url.String())
}

func (c Candidate) Delete(ctx context.Context) error {
	return localredis.Zdelete(ctx, "candidates", c.Url.String())
}
