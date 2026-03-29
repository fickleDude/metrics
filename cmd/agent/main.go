package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	cfg.ParseEnv()

	//инициализируем метрики
	service := agent.Init(cfg.RunAddr(), cfg.PollInterval(), cfg.ReportInterval())
	logger.Log.Debug("Инициализация...")
	// Запускаем горутину
	go service.Update(ctx)
	go service.Post(ctx)
	logger.Log.Debug("В работе")
	// Ждем сигнала
	<-sigChan
	logger.Log.Debug("\nПолучен Ctrl+Cзавершаем горутину...")
	cancel()                                   // Отменяем контекст
	time.Sleep(time.Duration(5) * time.Second) //ждем завершения горутин
	logger.Log.Debug("Программа завершена.")
}
