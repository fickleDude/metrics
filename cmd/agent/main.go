package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fickleDude/metrics.git/internal/agent"
	"github.com/fickleDude/metrics.git/internal/config"
)

func main() {
	//cancel context
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//agent
	cfg := config.GetConfig(config.Agent)
	agent := agent.Init(cfg.RunAddr(), cfg.PollInterval(), cfg.ReportInterval(), cfg.Key(), cfg.RateLimit())

	//jobs
	var wg sync.WaitGroup
	wg.Add(1)
	go agent.Post(ctx, &wg)

	wg.Add(1)
	go agent.Update(ctx, &wg)

	<-sigChan
	cancel()

	wg.Wait()
}
