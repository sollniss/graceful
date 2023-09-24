package main

import (
	"log"
	"net/http"
	"time"

	"github.com/sollniss/graceful"
)

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("start Hello World")
		time.Sleep(10 * time.Second)
		w.Write([]byte("Hello World"))
		log.Print("finish Hello World")
	})
	s := &http.Server{
		Handler: m,
		Addr:    "0.0.0.0:8080",
	}

	ctx := graceful.NotifyShutdown()
	err := graceful.ListenAndServe(ctx, s, 60*time.Second)
	if err != nil {
		log.Printf("error during shutdown: %s", err)
		return
	}
	log.Print("gracefully shut down")
}
