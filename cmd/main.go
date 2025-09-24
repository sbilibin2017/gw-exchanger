package main

import (
	"context"
	"log"

	"github.com/sbilibin2017/gw-exchanger/internal/app"
)

// main — точка входа в приложение.
func main() {
	// 0. Вывод информации о сборке в лог
	app.PrintBuildInfo()

	// 1. Парсинг командных флагов
	flags, err := app.ParseFlags()
	if err != nil {
		log.Fatalf("не удалось разобрать флаги: %v", err)
	}

	// 2. Загрузка конфигурации из файла или переменных окружения
	cfg, err := app.ParseConfig(flags.ConfigPath)
	if err != nil {
		log.Fatalf("не удалось загрузить конфигурацию: %v", err)
	}

	// 3. Запуск gRPC сервера с использованием конфигурации
	if err := app.Run(context.Background(), cfg); err != nil {
		log.Fatalf("сервер остановлен с ошибкой: %v", err)
	}

	// 4. Логирование успешного завершения работы
	log.Println("сервер остановлен корректно")
}
