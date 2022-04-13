package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/oleksandr/bonjour"
	"golang.org/x/net/context"
)

type Candidate struct {
	Address string
}

func NewCandidate(address string) Candidate {
	return Candidate{
		Address: address,
	}
}
func (c Candidate) Update(ctx context.Context) error {
	timestamp := time.Now().UTC().Unix()
	get := rdb.Get(ctx, "peer:by:"+c.Address)
	if err := get.Err(); err == nil {
		return errors.New("already exists as a peer")
	}

	intCmd := rdb.ZAdd(ctx, "candidates", &redis.Z{Score: float64(timestamp), Member: c.Address})
	return intCmd.Err()
}
func (c Candidate) Delete(ctx context.Context) error {
	intCmd := rdb.ZRem(ctx, "candidates", c.Address)

	return intCmd.Err()
}

type Peer struct {
	Id         string
	Address    string
	Candidates []string
}

func NewPeer(id string, address string) Peer {
	return Peer{
		Id:      id,
		Address: address,
	}
}
func (p Peer) Update(ctx context.Context) error {
	timestamp := time.Now().UTC().Unix()

	setExAddress := rdb.SetEX(ctx, "address:by:"+p.Id, p.Address, time.Second*30)
	if setExAddress.Err() != nil {
		log.Fatalln("Peer", p.Id, "Update", "address:by:"+p.Id, "error", setExAddress.Err())
	}
	setExPeer := rdb.SetEX(ctx, "peer:by:"+p.Address, p.Id, time.Second*30)
	if setExPeer.Err() != nil {
		log.Fatalln("Peer", p.Id, "Update", "peer:by:"+p.Address, "error", setExPeer.Err())
	}

	zAdd := rdb.ZAdd(ctx, "peers", &redis.Z{Score: float64(timestamp), Member: p.Id})
	if zAdd.Err() != nil {
		log.Fatalln("Peer", p.Id, "Update", "zadd", "error", zAdd.Err())
	}

	err := Candidate{
		Address: p.Address,
	}.Delete(ctx)
	if err != nil {
		// ignore, can be deleted by expiry
		log.Println("Peer", p.Id, "Update", "candidate", "delete error", err)
	}

	return nil
}

func (p Peer) Delete(ctx context.Context) error {
	delAddressBy := rdb.Del(ctx, "address:by:"+p.Id)
	if delAddressBy.Err() != nil {
		// ignore, can be deleted by expiry
		log.Println("Peer", p.Id, "Delete", "address:by:"+p.Id, "error", delAddressBy.Err())
	}
	delPeerBy := rdb.Del(ctx, "peer:by:"+p.Id)
	if delPeerBy.Err() != nil {
		// ignore, can be deleted by expiry
		log.Println("Peer", p.Id, "Delete", "peer:by:"+p.Id, "error", delPeerBy.Err())
	}

	zRem := rdb.ZRem(ctx, "peers", p.Id)
	if zRem.Err() != nil {
		// ignore, can be deleted by pruner
		log.Println("Peer", p.Id, "Delete", "peers", "error", zRem.Err())
	}
	return zRem.Err()
}

var rdb redis.Client
var id string

var port string
var teapots map[string]bool

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	id = uuid.New().String()

	teapots = make(map[string]bool)

	log.Println("USVA sierra ", id)

	rdb = *redis.NewClient(&redis.Options{})
	rdb.FlushAll(ctx)

	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	portInt, _ := strconv.Atoi(port)

	s, err := bonjour.Register("usvad sierra "+id, "_usva._tcp", "", portInt, []string{"txtv=1", "app=usvad-sierra-" + id}, nil)
	handler := make(chan os.Signal, 1)
	signal.Notify(handler, os.Interrupt)
	go func() {
		for sig := range handler {
			if sig == os.Interrupt {
				s.Shutdown()
				time.Sleep(1e9)
				break
			}
		}
	}()
	if err != nil {
		log.Fatalln(err.Error())
	}

	go server()
	time.Sleep(1 * time.Second)
	go discoverer(ctx)
	go pruner(ctx)

	<-ctx.Done()
}

func connect(ctx context.Context, peerAddress string) error {
	if teapots[peerAddress] {
		log.Println("would be a teapot", peerAddress)
		return nil
	}

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	query := "?host=" + peerAddress
	if os.Getenv("USVA_ADDRESS") != "" {
		query = query + "&address=" + os.Getenv("USVA_ADDRESS")
	}

	response, err := client.PostForm("http://"+peerAddress+"/.well-known/usva-sierra"+query, url.Values{
		"id": {id},
	})

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTeapot {
		teapots[peerAddress] = true
		return errors.New("teapot")
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("unexpected status: " + response.Status)
	}

	peer := Peer{}
	err = json.NewDecoder(response.Body).Decode(&peer)
	if err != nil {
		return errors.New("decode: " + err.Error())
	}

	peer.Address = peerAddress

	for _, candidateAddress := range peer.Candidates {
		err := Candidate{
			Address: candidateAddress,
		}.Update(ctx)

		if err != nil {
			log.Println("connect", "candidate", candidateAddress, "save error", err)
		} else {
			log.Println("connect", "candidate", candidateAddress, "save ok")
		}
	}

	err = peer.Update(ctx)
	if err != nil {
		log.Println("connect", "peer", peer.Id, peer.Address, "save error", err)
	} else {
		log.Println("connect", "peer", peer.Id, peer.Address, "save ok")
	}

	return nil
}
func peers(ctx context.Context) []string {
	zRangeByScore := rdb.ZRangeByScore(ctx, "peers", &redis.ZRangeBy{Min: "-Inf", Max: "+Inf"})
	if err := zRangeByScore.Err(); err != nil {
		log.Fatalln("err reply", err)
	}
	ids, err := zRangeByScore.Result()
	if err != nil {
		log.Fatalln("err result", err)
	}
	return ids
}

func candidates(ctx context.Context) []string {
	zRangeByScore := rdb.ZRangeByScore(ctx, "candidates", &redis.ZRangeBy{Min: "-Inf", Max: "+Inf"})
	if err := zRangeByScore.Err(); err != nil {
		log.Fatalln("err reply", err)
	}
	addresses, err := zRangeByScore.Result()
	if err != nil {
		log.Fatalln("err result", err)
	}
	return addresses
}

func discoverer(ctx context.Context) {
	seedString := "sierra.usva.io"
	if os.Getenv("USVA_SEEDS") != "" {
		seedString = os.Getenv("USVA_SEEDS")
	}

	seeds := strings.Split(seedString, ",")

	var bonjours = make(map[string]bool)
	for {
		resolver, err := bonjour.NewResolver(nil)
		if err != nil {
			log.Println("Failed to initialize resolver:", err.Error())
			os.Exit(1)
		}

		results := make(chan *bonjour.ServiceEntry)

		go func(results chan *bonjour.ServiceEntry) {
			for e := range results {
				log.Printf("BONJOUR", e.Instance, e.Service, e.AddrIPv4, e.Port, e.ServiceRecord, e.Text)
				bonjours[e.AddrIPv4.String()] = true
			}
		}(results)

		err = resolver.Browse("_usva._tcp", "local.", results)
		if err != nil {
			log.Println("Failed to browse:", err.Error())
		}

		for _, address := range candidates(ctx) {
			get := rdb.Get(ctx, "peer:by:"+address)
			if err := get.Err(); err == nil {
				// already a peer
				log.Fatalln("candidate", address, "already a peer")
			}

			log.Println("candidate", "connect", address)
			err := connect(ctx, address)
			if err != nil {
				log.Println("candidate connect error", err)
			}
			// // always delete candidate, it will come again
			// if err := candidate.Delete(ctx); err != nil {
			// 	log.Println("candidate", "delete", "error", err)
			// } else {
			// 	log.Println("candidate", "delete", "ok", candidate.Address)
			// }
		}

		peerIds := peers(ctx)
		log.Println("discover", "peerIds", peerIds)
		if len(peerIds) == 0 {
			log.Println("discover", "seeds", seeds)
			for _, seed := range seeds {
				if seed == "localhost" {
					continue
				}
				err := connect(ctx, seed)
				if err != nil {
					log.Println("seed connect err", err)
				}
			}
		}

		log.Println("discover", "bonjours", bonjours)
		for key := range bonjours {
			err := connect(ctx, key)
			if err != nil {
				log.Println("bonjour connect err", err)
			}
		}

		log.Println("discover", "peerIds", peerIds)
		for _, peerId := range peerIds {
			get := rdb.Get(ctx, "address:by:"+peerId)
			if err := get.Err(); err != nil {
				log.Println("peer", "address:by:"+peerId, "err", err)
				continue
			}
			address, err := get.Result()
			if err != nil {
				log.Println("peer", "address:by:"+peerId, "result", err)
				continue
			}
			log.Println("peer", "connecting", address)
			err = connect(ctx, address)
			if err != nil {
				log.Println("peer", "connect", address, "error", err)
				continue
			}
			log.Println("peer", "saved", address)
		}

		time.Sleep(time.Second * 5)
	}
}

func pruner(ctx context.Context) {
	for {
		ago := time.Now().UTC().Add(time.Second * -10).Unix()
		rdb.ZRemRangeByScore(ctx, "peers", "-Inf", strconv.FormatInt(ago, 30))
		rdb.ZRemRangeByScore(ctx, "candidates", "-Inf", strconv.FormatInt(ago, 30))
		time.Sleep(1 * time.Second)
	}
}

func server() {
	r := gin.Default()
	r.Use(location.Default())

	r.LoadHTMLGlob("./views/*")

	r.GET("/", func(c *gin.Context) {
		hostname := location.Get(c).String()
		if c.Request.Header.Get("X-Forwarded-Host") != "" {
			hostname = c.Request.Header.Get("X-Forwarded-Host")
		}
		// remove :80
		hostname = strings.Split(hostname, ":")[0]
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"id":       id,
			"hostname": hostname,
		})
	})

	r.GET("/app/peers", func(c *gin.Context) {
		c.HTML(http.StatusOK, "peers.tmpl", gin.H{})
	})
	r.GET("/app/peer/:id", func(c *gin.Context) {
		peerId := c.Param("id")
		c.HTML(http.StatusOK, "peer.tmpl", gin.H{
			"id": peerId,
		})
	})
	r.GET("/app/candidates", func(c *gin.Context) {
		c.HTML(http.StatusOK, "candidates.tmpl", gin.H{})
	})

	r.GET("/peers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"peers": peers(c),
		})
	})
	r.GET("/candidates", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"candidates": candidates(c),
		})
	})

	r.GET("/peer/:id", func(c *gin.Context) {
		peerId := c.Param("id")
		address, errAddress := rdb.Get(c, "address:by:"+peerId).Result()
		since, errScore := rdb.ZScore(c, "peers", peerId).Result()

		if errAddress != nil || errScore != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"id": peerId,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"id":      peerId,
				"address": address,
				"since":   since,
			})
		}
	})

	r.GET("/echo", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"clientIP": c.ClientIP(),
			"headers":  c.Request.Header,
		})
	})

	r.POST("/.well-known/usva-sierra", func(c *gin.Context) {
		peerId := c.PostForm("id")
		if peerId == id {
			c.JSON(http.StatusTeapot, gin.H{})
			return
		}

		peerAddress := c.ClientIP()
		if c.Request.Header.Get("X-Original-Forwarded-For") != "" {
			peerAddress = c.Request.Header.Get("X-Original-Forwarded-For")
		}
		if c.Query("address") != "" {
			peerAddress = c.Query("address")
		}

		err := Candidate{
			Address: peerAddress,
		}.Update(c)
		if err != nil {
			log.Println("/.well-known", "Candidate", "Save", "error", err)
		} else {
			log.Println("/.well-known", "Candidate", "Save", "ok", peerAddress)

		}

		c.JSON(http.StatusOK, gin.H{
			"id":         id,
			"candidates": candidates(c),
		})
	})

	r.Run("0.0.0.0:" + port)
}
