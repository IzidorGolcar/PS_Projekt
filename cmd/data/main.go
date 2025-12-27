package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"seminarska/internal/data"
	"seminarska/internal/data/config"
	"time"
)

func main() {
	cfg := config.Load()
	configureLogger(cfg)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	log.Println("Starting data service")
	service := data.NewService(ctx, cfg)
	<-ctx.Done()
	log.Println("Stopping data service")
	select {
	case <-time.After(30 * time.Second):
		log.Println("Forcefully stopping data service")
	case <-service.Done():
		log.Println("Data service stopped")
	}
}

func configureLogger(cfg config.NodeConfig) {
	if cfg.LogPath == "" {
		log.Default().SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal(err)
		}
		log.Default().SetOutput(file)
	}

	log.Default().SetFlags(log.LstdFlags | log.Lmsgprefix)
	prefix := fmt.Sprintf("[%s] ", cfg.NodeId)
	log.Default().SetPrefix(prefix)
}
