package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fickleDude/metrics.git/internal/agent"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Настраиваем отслеживание SIGINT (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//инициализируем метрики
	service := agent.Init(2, 2)
	fmt.Println("Инициализация...")
	// Запускаем горутину
	// for i := 0; i < 3; i++ {
	go service.Update(ctx)
	go service.Post(ctx)

	//}
	fmt.Println("В работе")
	// Ждем сигнала
	<-sigChan
	fmt.Println("\nПолучен Ctrl+Cзавершаем горутину...")
	cancel() // Отменяем контекст
	fmt.Println("Программа завершена.")
}
