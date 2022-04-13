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

type Candidate struct {
	Address string
}

func NewCandidate(address string) Candidate {
	return Candidate{
		Address: address,
	}
}
func (p Candidate) Save(ctx context.Context) error {
	timestamp := time.Now().UTC().Unix()
	intCmd := rdb.ZAdd(ctx, "candidates", &redis.Z{Score: float64(timestamp), Member: p.Address})

	return intCmd.Err()
}
func (p Candidate) Delete(ctx context.Context) error {
	intCmd := rdb.ZRem(ctx, "candidate", p.Address)

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
func (p Peer) Save(ctx context.Context) error {
	timestamp := time.Now().UTC().Unix()

	setEx := rdb.SetEX(ctx, "address:"+p.Id, p.Address, time.Second*30)
	if setEx.Err() != nil {
		log.Println("Peer", p.Id, "Save", "address:"+p.Id, "error", setEx.Err())
	}
	zAdd := rdb.ZAdd(ctx, "peers", &redis.Z{Score: float64(timestamp), Member: p.Id})

	return zAdd.Err()
}
func (p Peer) Delete(ctx context.Context) error {
	delAddress := rdb.Del(ctx, "address:"+p.Id)
	if delAddress.Err() != nil {
		// ignore, can be deleted by expiry
		log.Println("Peer", p.Id, "Delete", "address:"+p.Id, "error", delAddress.Err())
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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	id = uuid.New().String()
	log.Println("USVA GALANT ", id)

	rdb = *redis.NewClient(&redis.Options{})

	go discoverer(ctx)
	go pruner(ctx)
	go server()

	<-ctx.Done()
}

func connect(ctx context.Context, peerAddress string) error {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	query := "?host=" + peerAddress
	if os.Getenv("USVA_ADDRESS") != "" {
		query = query + "&address=" + os.Getenv("USVA_ADDRESS")
	}
	response, err := client.PostForm("http://"+peerAddress+"/.well-known/usva-galant"+query, url.Values{
		"id": {id},
	})

	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusTeapot {
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

	for _, peerAddress := range peer.Candidates {
		err := Candidate{
			Address: peerAddress,
		}.Save(ctx)

		if err != nil {
			log.Println("connect", "candidate", peerAddress, "save error", err)
		}
	}
	return peer.Save(ctx)
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
	seedString := "galant.usva.io"
	if os.Getenv("USVA_SEEDS") != "" {
		seedString = os.Getenv("USVA_SEEDS")
	}

	seeds := strings.Split(seedString, ",")

	for {
		for _, address := range candidates(ctx) {
			Candidate := Candidate{
				Address: address,
			}
			err := connect(ctx, Candidate.Address)
			if err != nil {
				log.Println("candidate connect error", err)
				Candidate.Delete(ctx)
				continue
			}
		}

		peerIds := peers(ctx)
		if len(peerIds) == 0 {
			for _, seed := range seeds {
				err := connect(ctx, seed)
				if err != nil {
					log.Println("seed connect err", err)
				}
			}
		}

		for _, peerId := range peerIds {
			get := rdb.Get(ctx, "address:"+peerId)
			if err := get.Err(); err != nil {
				log.Println("get address err", peerId, err)
				continue
			}
			address, err := get.Result()
			if err != nil {
				log.Println("get address result err", err)
				continue
			}
			err = connect(ctx, address)
			if err != nil {
				log.Println("peer connect error", err)
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

	r.POST("/.well-known/usva-galant", func(c *gin.Context) {
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

		Candidate{
			Address: peerAddress,
		}.Save(c)

		c.JSON(http.StatusOK, gin.H{
			"id":         id,
			"candidates": candidates(c),
		})
	})

	r.Run("0.0.0.0:8080")
}
