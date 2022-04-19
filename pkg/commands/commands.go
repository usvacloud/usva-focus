package commands

import (
	"context"
	"log"

	"github.com/usvacloud/usva-focus/pkg/types"
)

func Peers(ctx context.Context) {
	for _, peer := range types.Peers(ctx, 100) {
		log.Println("peer", peer)
	}
}
func Peer(ctx context.Context) {

	url := types.GetRandomPeer(ctx)
	log.Println(url)
	// client := http.Client{
	// 	Timeout: 3 * time.Second,
	// }

	// query := "?host=" + peerAddress
	// if os.Getenv("USVA_ADDRESS") != "" {
	// 	query = query + "&address=" + os.Getenv("USVA_ADDRESS")
	// }

	// response, err := client.PostForm("http://"+peerAddress+"/.well-known/usva-focus"+query, url.Values{
	// 	"id": {id},
	// })

}
