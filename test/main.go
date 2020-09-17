package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/t0mk/logdna"
)

func main() {
	cfg := logdna.Config{
		APIKey:   os.Getenv("LOGDNA_API_KEY"),
		Hostname: "dev.parce.ly",
		Env:      "void",
		App:      "logdna-test",
	}
	if cfg.APIKey == "" {
		panic("empoty api key")
	}
	c := logdna.NewClient(cfg)
	c.Run()
	c.Err("Hi", time.Now().UTC().String())

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println("received signal", <-ch)
	log.Println("quitting no matter what")

}
