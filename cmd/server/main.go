package main

import (
	_ "log/slog"
	"os"

	"pvz-service-avito-internship/internal/app"
	"pvz-service-avito-internship/internal/config"
	"pvz-service-avito-internship/pkg/logger"
)

func main() {
	// 1. Загрузка конфигурации
	cfg := config.Load()

	// 2. Инициализация логгера
	log := logger.Setup(cfg.Logger.Level)

	// 3. Создание экземпляра приложения с обработкой паники при инициализации
	application := app.MustNewApp(cfg, log)

	// 4. Запуск приложения (включая graceful shutdown)
	application.Run()

	log.Info("Application finished")
	os.Exit(0) // Успешный выход
}
