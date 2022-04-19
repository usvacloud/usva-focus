package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/usvacloud/usva-focus/pkg/types"
)

func Routes(router *gin.RouterGroup) {
	router.PUT("candidate", UpdateCandidate)
	router.GET("candidates", Candidates)
	router.GET("peers", Peers)
	router.GET("peer", Peer)
}

func UpdateCandidate(c *gin.Context) {
	var candidateHost string
	if a := c.PostForm("address"); a != "" {
		candidateHost = a
	} else if a := c.Request.Header.Get("X-Original-Forwarded-For"); a != "" {
		candidateHost = a
	} else {
		candidateHost = c.ClientIP()
	}
	var candidatePort string
	if p := c.PostForm("port"); p != "" {
		candidatePort = p
	} else {
		candidatePort = "80"
	}

	candidateUrl := &url.URL{
		Scheme: "http",
		Host:   candidateHost + ":" + candidatePort,
	}

	err := types.NewCandidate(candidateUrl.String()).Update(c)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func Candidates(c *gin.Context) {
	urlStrings := []string{}
	for _, candidate := range types.Candidates(c, 100) {
		urlStrings = append(urlStrings, candidate.Url.String())
	}

	c.JSON(http.StatusOK, gin.H{
		"candidates": urlStrings,
	})
}

func Peers(c *gin.Context) {
	urlStrings := []string{}
	for _, peer := range types.Peers(c, 100) {
		urlStrings = append(urlStrings, peer.Url.String())
	}
	c.JSON(http.StatusOK, gin.H{
		"peers": urlStrings,
	})
}

func Peer(c *gin.Context) {
	peerId := c.Query("id")
	since, err := types.NewPeer(peerId).Since(c)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"peer":  peerId,
			"since": since,
		})
	}
}
