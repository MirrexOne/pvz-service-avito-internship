package main

import (
	_ "log/slog"
	"os"

	"pvz-service-avito-internship/internal/app"
	"pvz-service-avito-internship/internal/config"
	"pvz-service-avito-internship/pkg/logger"
)

func main() {
	cfg := config.Load()

	log := logger.Setup(cfg.Logger.Level)

	application := app.MustNewApp(cfg, log)

	application.Run()

	log.Info("Application finished")
	os.Exit(0)
}
