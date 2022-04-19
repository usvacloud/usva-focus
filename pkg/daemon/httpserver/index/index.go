package index

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/usvacloud/usva-focus/pkg/usva"
)

func Routes(router *gin.RouterGroup) {
	router.GET("", Index)
	router.GET("healthz", Healthz)
	router.GET(".well-known/usva", WellKnownUsva)
}

func Index(c *gin.Context) {
	hostname := location.Get(c).String()
	if c.Request.Header.Get("X-Forwarded-Host") != "" {
		hostname = c.Request.Header.Get("X-Forwarded-Host")
	}
	// remove :80
	hostname = strings.Split(hostname, ":")[0]
	c.HTML(http.StatusOK, "index", gin.H{
		"id":       usva.Id,
		"hostname": hostname,
	})
}

func Healthz(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func WellKnownUsva(c *gin.Context) {
	// var candidateAddress string
	// if c.Query("address") != "" {
	// 	candidateAddress = c.Query("address")
	// } else if c.Request.Header.Get("X-Original-Forwarded-For") != "" {
	// 	candidateAddress = c.Request.Header.Get("X-Original-Forwarded-For")
	// } else {
	// 	candidateAddress = c.ClientIP()
	// }

	// err := Candidate{
	// 	Address: candidateAddress,
	// }.Update(c)
	// if err != nil {
	// 	log.Println("/.well-known", "Candidate", "Save", "error", err)
	// } else {
	// 	log.Println("/.well-known", "Candidate", "Save", "ok", peerAddress)

	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"id":         id,
	// 	"candidates": candidates(c),
	// })

	c.JSON(http.StatusOK, gin.H{
		"id":    usva.Id,
		"model": usva.Model,
	})
}
