package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fickleDude/metrics.git/internal/agent"
	"github.com/fickleDude/metrics.git/internal/config"
	"github.com/fickleDude/metrics.git/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Настраиваем отслеживание SIGINT (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//config
	cfg := config.NewConfig()
	cfg.ParseFlags("agent")
	cfg.ParseEnv("agent")

	//инициализируем метрики
	service := agent.Init(cfg.RunAddr(), cfg.PollInterval(), cfg.ReportInterval())
	logger.Log.Debug("Инициализация...")
	// Запускаем горутину
	var wg sync.WaitGroup
	wg.Add(2)
	go service.Update(ctx, &wg)
	go service.Post(ctx, &wg)
	logger.Log.Debug("В работе")
	// Ждем сигнала
	<-sigChan
	logger.Log.Debug("\nПолучен Ctrl+Cзавершаем горутину...")
	cancel() // Отменяем контекст
	wg.Wait()
	logger.Log.Debug("Программа завершена.")
}
