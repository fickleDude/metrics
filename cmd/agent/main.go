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
	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Настраиваем отслеживание SIGINT (Ctrl+C)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//инициализируем метрики
	service := agent.Init(2, 10)

	// Запускаем горутину
	// for i := 0; i < 3; i++ {
	go service.Post(ctx)
	go service.Update(ctx)
	//}

	// Ждем сигнала
	<-sigChan
	fmt.Println("\nПолучен Ctrl+Cзавершаем горутину...")
	cancel() // Отменяем контекст
	fmt.Println("Программа завершена.")
}
