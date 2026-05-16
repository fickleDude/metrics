package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fickleDude/metrics.git/internal/agent"
	"github.com/fickleDude/metrics.git/internal/config"
	"golang.org/x/sync/errgroup"
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
	g := new(errgroup.Group)
	for w := 0; w <= 100; w++ {
		wg.Add(1)
		g.Go(func() error {
			return agent.Post(ctx, w, &wg)
		})
	}

	wg.Add(1)
	go agent.Update(ctx, &wg)

	<-sigChan
	cancel()

	if err := g.Wait(); err != nil {
		log.Println(err)
	}

	wg.Wait()
}
