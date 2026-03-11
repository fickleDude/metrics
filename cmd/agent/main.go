package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fickleDude/metrics.git/internal/agent"
	"github.com/fickleDude/metrics.git/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Настраиваем отслеживание SIGINT (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//config
	cfg := config.NewConfig()
	cfg.ParseFlags()
	cfg.ParseEnv()

	//инициализируем метрики
	service := agent.Init(cfg.RunAddr(), cfg.PollInterval(), cfg.ReportInterval())
	fmt.Println("Инициализация...")
	// Запускаем горутину
	go service.Update(ctx)
	go service.Post(ctx)
	fmt.Println("В работе")
	// Ждем сигнала
	<-sigChan
	fmt.Println("\nПолучен Ctrl+Cзавершаем горутину...")
	cancel()                                   // Отменяем контекст
	time.Sleep(time.Duration(5) * time.Second) //ждем завершения горутин
	fmt.Println("Программа завершена.")
}
