package types

import (
	"context"
	"errors"
	"log"
	"net/url"
	"time"

	"github.com/usvacloud/usva-focus/pkg/localredis"
)

type Self struct {
	Url *url.URL
}

func NewSelf(urlString string) *Self {
	u, err := url.Parse(urlString)
	if err != nil {
		log.Fatalln("url", err)
	}
	return &Self{
		Url: u,
	}
}
func (s Self) Since(ctx context.Context) (float64, error) {
	now := time.Now().UTC().Unix()

	score, err := localredis.Zscore(ctx, "selfs", s.Url.String())
	if err != nil {
		return -1, errors.New("not found")
	}
	delta := float64(now) - score
	return delta, nil
}

func (s Self) Update(ctx context.Context) error {
	return localredis.Zupdate(ctx, "selfs", s.Url.String())
}
