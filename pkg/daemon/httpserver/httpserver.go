package httpserver

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"

	"github.com/usvacloud/usva-focus/pkg/daemon/httpserver/api"
	"github.com/usvacloud/usva-focus/pkg/daemon/httpserver/app"
	"github.com/usvacloud/usva-focus/pkg/daemon/httpserver/index"
	"github.com/usvacloud/usva-focus/pkg/usva"
)

func Run(ctx context.Context) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(
		location.Default(),
	)
	r.LoadHTMLGlob("./templates/**/*.gohtml")

	api.Routes(r.Group("/api"))
	app.Routes(r.Group("/app"))
	index.Routes(r.Group("/"))

	var address string
	if port, ok := os.LookupEnv("PORT"); ok {
		address = ":" + port
	} else {
		address = ":0"
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	go func() {
		err = http.Serve(listener, r)
		if err != nil {
			log.Println("http.Serve", "Error", err)
		}
	}()

	usva.SetPort(listener.Addr().(*net.TCPAddr).Port)

	<-ctx.Done()
	log.Println("http-server shutdown started")

	log.Println("http-server did shut down")
}
