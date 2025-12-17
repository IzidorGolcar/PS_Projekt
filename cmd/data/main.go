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
	configureLogger(cfg.NodeId)
	ctx, cancel := signal.NotifyContext(context.Background())
	defer cancel()
	log.Println("Starting data service")
	service := data.NewService(ctx, cfg)
	<-ctx.Done()
	log.Println("Stopping data service")
	select {
	case <-time.After(5 * time.Second):
		log.Println("Forcefully stopping data service")
	case <-service.Done():
		log.Println("Data service stopped")
	}
}

func configureLogger(serviceId string) {
	log.Default().SetOutput(os.Stdout)
	log.Default().SetFlags(log.LstdFlags | log.Lmsgprefix)
	prefix := fmt.Sprintf("[%s] ", serviceId)
	log.Default().SetPrefix(prefix)
}
