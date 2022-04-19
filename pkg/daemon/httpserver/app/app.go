package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routes(router *gin.RouterGroup) {
	router.GET("peers", Peers)
	router.GET("peer", Peer)
	router.GET("candidates", Candidates)
}

func Peers(c *gin.Context) {
	c.HTML(http.StatusOK, "app/peers", gin.H{})
}

func Peer(c *gin.Context) {
	peerId := c.Query("id")
	c.HTML(http.StatusOK, "app/peer", gin.H{
		"id": peerId,
	})
}
func Candidates(c *gin.Context) {
	c.HTML(http.StatusOK, "app/candidates", gin.H{})
}
