package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config определяет общую структуру конфигурации всего приложения.
type Config struct {
	HTTPServer   `yaml:"http"`     // Конфигурация основного HTTP сервера
	GRPCServer   `yaml:"grpc"`     // Конфигурация gRPC сервера
	Metrics      `yaml:"metrics"`  // Конфигурация сервера метрик Prometheus
	Database     `yaml:"database"` // Конфигурация основной базы данных
	Auth         `yaml:"auth"`     // Конфигурация аутентификации и JWT
	Logger       `yaml:"logger"`   // Конфигурация логгера
	Hasher       `yaml:"hasher"`   // Конфигурация хэшера паролей
	TestDatabase Database          `yaml:"test_database"` // Конфигурация тестовой базы данных (используется только в тестах)
}

// HTTPServer содержит настройки для основного HTTP сервера.
type HTTPServer struct {
	// Port - порт, на котором будет слушать HTTP сервер.
	Port string `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
}

// GRPCServer содержит настройки для gRPC сервера.
type GRPCServer struct {
	// Port - порт, на котором будет слушать gRPC сервер.
	Port string `yaml:"port" env:"GRPC_PORT" env-default:"3000"`
}

// Metrics содержит настройки для сервера метрик Prometheus.
type Metrics struct {
	// Port - порт, на котором будет слушать сервер метрик (/metrics).
	Port string `yaml:"port" env:"METRICS_PORT" env-default:"9000"`
}

// Database содержит настройки для подключения к базе данных PostgreSQL.
type Database struct {
	Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"DB_USER" env-default:"user"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-default:"password"`
	Name     string `yaml:"name" env:"DB_NAME" env-default:"pvz_db"`
	SSLMode  string `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
}

type Auth struct {
	JWTSecret string        `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
	JWTttl    time.Duration `yaml:"jwt_ttl_hours" env:"JWT_TTL_HOURS" env-default:"24h"`
}

// Logger содержит настройки для логгера приложения.
type Logger struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

// Hasher содержит настройки для хэширования паролей.
type Hasher struct {
	// BcryptCost - определяет вычислительную сложность (стоимость) хеширования bcrypt.
	// Чем выше значение, тем безопаснее, но медленнее.
	BcryptCost int `yaml:"bcrypt_cost" env:"BCRYPT_COST" env-default:"10"`
}

// Load загружает конфигурацию приложения.
// Порядок приоритета:
// 1. Переменные окружения (самый высокий приоритет).
// 2. Значения из YAML файла (если найден).
// 3. Значения по умолчанию (env-default).
// Функция паникует, если не удается прочитать обязательные переменные окружения (env-required).
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("INFO: .env file not found or error loading it: %v. Relying on existing environment variables.", err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yml"
	}

	var cfg Config

	if _, err := os.Stat(configPath); err == nil {
		err := cleanenv.ReadConfig(configPath, &cfg)
		if err != nil {
			log.Printf("WARN: Error reading config file '%s': %v. Relying solely on environment variables.", configPath, err)
		} else {
			log.Printf("INFO: Loaded base configuration structure from file: %s", configPath)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("WARN: Error accessing config file '%s': %v. Relying solely on environment variables.", configPath, err)
	} else {
		log.Printf("INFO: Configuration file not found at '%s'. Relying solely on environment variables.", configPath)
	}

	log.Printf("INFO: Reading environment variables (will override YAML values if any)...")
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("FATAL: Error reading environment variables: %v", err)
	}

	log.Printf("INFO: Configuration loaded successfully. Log Level: %s, HTTP Port: %s, gRPC Port: %s",
		cfg.Logger.Level, cfg.HTTPServer.Port, cfg.GRPCServer.Port)

	return &cfg
}

func LoadTestConfig() *Config {
	cfg := Load()

	if host := os.Getenv("TEST_DB_HOST"); host != "" {
		cfg.TestDatabase.Host = host
	}
	if port := os.Getenv("TEST_DB_PORT_HOST"); port != "" { // Порт для доступа к тестовой БД: 5433 доступ с хоста. Требуется для того, чтобы запускать тесты локально
		cfg.TestDatabase.Port = port
	}
	if user := os.Getenv("TEST_DB_USER"); user != "" {
		cfg.TestDatabase.User = user
	}
	if password := os.Getenv("TEST_DB_PASSWORD"); password != "" {
		cfg.TestDatabase.Password = password
	}
	if name := os.Getenv("TEST_DB_NAME"); name != "" {
		cfg.TestDatabase.Name = name
	}
	if sslMode := os.Getenv("TEST_DB_SSL_MODE"); sslMode != "" {
		cfg.TestDatabase.SSLMode = sslMode
	}

	log.Printf("INFO: Test database configuration loaded. Host: %s, Port: %s, Name: %s",
		cfg.TestDatabase.Host, cfg.TestDatabase.Port, cfg.TestDatabase.Name)

	return cfg
}
