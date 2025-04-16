package logger

import (
	"log/slog"
	"os"
	"strings" // Добавлен для обработки уровня лога
)

// Setup настраивает и возвращает экземпляр логгера slog.
// Принимает уровень логирования ("debug", "info", "warn", "error").
// Использует JSONHandler для вывода логов в формате JSON в stdout.
func Setup(level string) *slog.Logger {
	var logLevel slog.Level

	// Определяем уровень логирования на основе строки конфигурации
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		// Если указан неверный уровень, используем Info по умолчанию
		logLevel = slog.LevelInfo
		slog.Warn("Invalid log level specified, using default level: info", slog.String("invalid_level", level))
	}

	// Настраиваем опции обработчика slog
	opts := &slog.HandlerOptions{
		Level: logLevel,
		// AddSource: true, // Раскомментируйте, чтобы добавить имя файла и номер строки в лог (может влиять на производительность)
	}

	// Создаем JSON обработчик, который пишет в стандартный вывод (stdout)
	handler := slog.NewJSONHandler(os.Stdout, opts)

	// Создаем новый логгер с настроенным обработчиком
	logger := slog.New(handler)

	// Устанавливаем созданный логгер как логгер по умолчанию для пакета log/slog (опционально)
	// slog.SetDefault(logger)

	logger.Info("Logger initialized", slog.String("level", logLevel.String()))

	return logger
}
