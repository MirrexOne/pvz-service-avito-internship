package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

const (
	defaultMaxPoolSize       = 25
	defaultMinPoolSize       = 5
	defaultMaxConnIdleTime   = 5 * time.Minute
	defaultMaxConnLifetime   = 1 * time.Hour
	defaultHealthCheckPeriod = 1 * time.Minute
	connectTimeout           = 5 * time.Second
)

// NewPostgresPool создает и настраивает пул соединений к PostgreSQL.
func NewPostgresPool(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database DSN: %w", err)
	}

	// Настройки пула соединений
	config.MaxConns = defaultMaxPoolSize
	config.MinConns = defaultMinPoolSize
	config.MaxConnIdleTime = defaultMaxConnIdleTime
	config.MaxConnLifetime = defaultMaxConnLifetime
	config.HealthCheckPeriod = defaultHealthCheckPeriod

	// Устанавливаем таймаут на подключение
	connCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connCtx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Проверяем соединение сразу после создания пула
	pingCtx, pingCancel := context.WithTimeout(ctx, connectTimeout)
	defer pingCancel()
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close() // Важно закрыть пул, если пинг не удался
		return nil, fmt.Errorf("unable to ping database after connect: %w", err)
	}

	logger.Info("Successfully connected to database")
	return pool, nil
}

// BuildDSN собирает строку Data Source Name (DSN) для подключения к PostgreSQL.
func BuildDSN(cfgHost, cfgPort, cfgUser, cfgPassword, cfgName, cfgSSLMode string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfgHost, cfgPort, cfgUser, cfgPassword, cfgName, cfgSSLMode)
}

func hidePassword(dsn string) string {
	// Простой поиск и замена, можно улучшить регулярным выражением
	passwordPrefix := "password="
	startIndex := strings.Index(dsn, passwordPrefix)
	if startIndex == -1 {
		return dsn // Пароль не найден
	}
	startIndex += len(passwordPrefix)
	endIndex := strings.Index(dsn[startIndex:], " ") // Ищем следующий пробел
	if endIndex == -1 {
		// Если пароль - последнее значение
		return dsn[:startIndex] + "********"
	}
	return dsn[:startIndex] + "********" + dsn[startIndex+endIndex:]
}
