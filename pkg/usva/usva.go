package usva

import (
	"context"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

var Id string
var Model string
var Port string
var PortInt int
var DaemonStarted chan bool

var once sync.Once

func Initialize() {
	once.Do(func() {
		Id = uuid.NewString()
		Model = "focus"
		DaemonStarted = make(chan bool)
	})
}

func SetPort(portInt int) {
	port := strconv.Itoa(portInt)

	Port = port
	PortInt = portInt
	close(DaemonStarted)
}

func Wait(ctx context.Context, channel chan bool) {
	select {
	case <-ctx.Done():
		return
	case <-channel:
		return
	}
}

type VoidFn func()

func SpawnVoidFn(wg *sync.WaitGroup, f VoidFn) {
	wg.Add(1)
	go func() {
		f()
		wg.Done()
	}()
}
