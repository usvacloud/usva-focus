package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Peer struct {
	Id        string
	Address   string
	Addresses []string
}

var rdb redis.Client
var id string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	id = uuid.New().String()
	log.Println("USVA FIESTA ", id)
	rdb = *redis.NewClient(&redis.Options{})

	go discoverer(ctx)
	go pruner(ctx)
	go server()

	<-ctx.Done()
}
func NewPeer(id string, address string) Peer {
	return Peer{
		Id:      id,
		Address: address,
	}
}
func (p Peer) Save(ctx context.Context) error {
	timestamp := time.Now().UTC().Unix()

	statusCmd := rdb.SetEX(ctx, "address:"+p.Id, p.Address, time.Second*30)
	if err := statusCmd.Err(); err != nil {
		return err
	}
	intCmd := rdb.ZAdd(ctx, "peers", &redis.Z{Score: float64(timestamp), Member: p.Id})

	return intCmd.Err()
}
func (p Peer) Delete(ctx context.Context) error {
	rdb.Del(ctx, "address:"+p.Id)
	rdb.ZRem(ctx, "peers", p.Id)

	return nil
}

func connect(ctx context.Context, peerAddress string) error {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	query := "?host=" + peerAddress
	if os.Getenv("USVA_ADDRESS") != "" {
		query = query + "&address=" + os.Getenv("USVA_ADDRESS")
	}
	response, err := client.PostForm("http://"+peerAddress+"/.well-known/usva-fiesta"+query, url.Values{
		"id":    {id},
		"peers": peers(ctx),
	})

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTeapot {
		return nil
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("unexpected status: " + response.Status)
	}

	peer := Peer{}
	err = json.NewDecoder(response.Body).Decode(&peer)
	if err != nil {
		return errors.New("decode: " + err.Error())
	}
	if peer.Id == id {
		return nil
	}

	peer.Address = peerAddress

	for _, peerAddress := range peer.Addresses {
		NewPeer(uuid.NewString(), peerAddress).Save(ctx)
	}
	return peer.Save(ctx)
}
func peers(ctx context.Context) []string {
	peersReply := rdb.ZRangeByScore(ctx, "peers", &redis.ZRangeBy{Min: "-Inf", Max: "+Inf"})
	if err := peersReply.Err(); err != nil {
		log.Fatalln("err reply", err)
	}
	peers, err := peersReply.Result()
	if err != nil {
		log.Fatalln("err result", err)
	}
	return peers
}

func addresses(ctx context.Context) []string {
	addresses := []string{}

	for _, peerId := range peers(ctx) {
		reply := rdb.Get(ctx, "address:"+peerId)
		if reply.Err() != nil {
			continue
		}
		address, err := reply.Result()
		if err != nil {
			continue
		}
		addresses = append(addresses, address)
	}

	return addresses
}
func discoverer(ctx context.Context) {
	seedString := "fiesta.usva.io"
	if os.Getenv("USVA_SEEDS") != "" {
		seedString = os.Getenv("USVA_SEEDS")
	}

	seeds := strings.Split(seedString, ",")

	for {
		stringSliceCmd := rdb.ZRangeByScore(ctx, "peers", &redis.ZRangeBy{Min: "-Inf", Max: "+Inf"})
		if err := stringSliceCmd.Err(); err != nil {
			log.Fatalln("peers err1", err)
		}
		peerIds, err := stringSliceCmd.Result()
		if err != nil {
			log.Fatalln("peers err2", err)
		}

		if len(peerIds) == 0 {
			for _, seed := range seeds {
				err := connect(ctx, seed)
				if err != nil {
					log.Println("seed connect err", err)
				}
			}
		}

		for _, peerId := range peerIds {
			peer := NewPeer(peerId, "")

			reply := rdb.Get(ctx, "address:"+peerId)
			if err := reply.Err(); err != nil {
				log.Println("get address err", peerId, err)
				peer.Delete(ctx)
				continue
			}
			peerAddress, err := reply.Result()
			if err != nil {
				log.Println("get address result err", err)
				peer.Delete(ctx)
				continue
			}

			err = connect(ctx, peerAddress)
			if err != nil {
				log.Println("peer connect error", err)
				peer.Delete(ctx)
				continue
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func pruner(ctx context.Context) {
	for {
		ago := time.Now().UTC().Add(time.Second * -10).Unix()
		rdb.ZRemRangeByScore(ctx, "peers", "-Inf", strconv.FormatInt(ago, 30))
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

	r.GET("/peers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"peers": peers(c),
		})
	})
	r.GET("/addresses", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"addresses": addresses(c),
		})
	})

	r.GET("/peer/:id", func(c *gin.Context) {
		peerId := c.Param("id")
		address, errAddress := rdb.Get(c, "address:"+peerId).Result()
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

	r.POST("/.well-known/usva-fiesta", func(c *gin.Context) {
		peerId := c.PostForm("id")
		peerPeerAddresses := strings.Split(c.PostForm("addresses"), ",")
		for _, peerPeerAddress := range peerPeerAddresses {
			peerPeer := Peer{
				Id:      uuid.NewString(),
				Address: peerPeerAddress,
			}
			peerPeer.Save(c)
		}

		peerAddress := c.ClientIP()
		if c.Request.Header.Get("X-Original-Forwarded-For") != "" {
			peerAddress = c.Request.Header.Get("X-Original-Forwarded-For")
		}
		if c.Query("address") != "" {
			peerAddress = c.Query("address")
		}

		peer := Peer{
			Id:      peerId,
			Address: peerAddress,
		}

		if peerId == id {
			c.JSON(http.StatusTeapot, gin.H{})
			peer.Delete(c)
			return
		}

		peer.Save(c)

		c.JSON(http.StatusOK, gin.H{
			"id":        id,
			"addresses": addresses(c),
		})
	})

	r.Run("0.0.0.0:8080")
}
