package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/apg/logflect"
)

func awaitSignals(cs ...io.Closer) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigCh
	log.Printf("Got signal: %q", sig)
	for _, c := range cs {
		c.Close()
	}
}

func main() {
	httpServer := &http.Server{Addr: ":9000"}
	shutdownChan := make(chan struct{})
	store := logflect.NewStore(logflect.MaxFeedAge, logflect.MaxSessionAge)
	server := logflect.NewServer(httpServer, store, shutdownChan)

	go awaitSignals(server)
	go server.Shutdown()
	server.Run()
}
