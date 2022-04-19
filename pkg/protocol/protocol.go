package protocol

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/usvacloud/usva-focus/pkg/types"
	"github.com/usvacloud/usva-focus/pkg/usva"
)

func Connect(ctx context.Context, url *url.URL) {
	self := types.NewSelf(url.String())
	if since, err := self.Since(ctx); err == nil {
		log.Println("self", url.String(), "since", since)
		return
	}
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	url.Path = "/.well-known/usva"
	log.Println("connect", url.String())

	response, err := client.Get(url.String())
	if err != nil {
		log.Println("usva", "connect", "err", err)
		return
	}

	if response.StatusCode != http.StatusOK {
		log.Println("usva", "connect", "status", response.StatusCode)
		return
	}

	var result map[string]interface{}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("usva", "connect", "read", err)
	}

	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		log.Println("usva", "connect", "json", err)
	}

	peerId := result["id"]
	if peerId == usva.Id {
		self.Update(ctx)
		return
	}

	url.Path = ""
	types.NewPeer(url.String()).Update(ctx)
}
