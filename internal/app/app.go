package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Регистрирует драйвер для postgresql:// DSN
	_ "github.com/golang-migrate/migrate/v4/source/file"       // Регистрирует драйвер для file:// путей
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq" // Драйвер pq часто является неявной зависимостью драйвера postgres migrate
	"google.golang.org/grpc"
	_ "log"

	// Используем относительные пути
	"pvz-service-avito-internship/internal/config"
	"pvz-service-avito-internship/internal/domain"
	httpHandler "pvz-service-avito-internship/internal/handler/http"
	promMetrics "pvz-service-avito-internship/internal/metrics"
	mw "pvz-service-avito-internship/internal/middleware"
	"pvz-service-avito-internship/internal/repository/postgres"
	"pvz-service-avito-internship/internal/service"
	grpcTransport "pvz-service-avito-internship/internal/transport/grpc"
	"pvz-service-avito-internship/pkg/database"
	"pvz-service-avito-internship/pkg/hash"
)

type App struct {
	cfg           *config.Config
	log           *slog.Logger
	dbPool        *pgxpool.Pool
	router        *gin.Engine
	server        *http.Server
	grpcServer    *grpcTransport.Server
	metricsServer *http.Server
}

// MustNewApp создает приложение, подключается к БД и применяет миграции.
// Паникует при ошибке подключения к БД или ошибке миграций.
func MustNewApp(cfg *config.Config, log *slog.Logger) *App {
	const op = "app.MustNewApp"
	log = log.With(slog.String("op", op))

	// --- 1. Подключение к основной базе данных ---
	log.Info("Attempting to connect to the main database...")
	dbDSN := database.BuildDSN(
		cfg.Database.Host, // При локальном запуске должен быть 'localhost', в Docker 'db' (из .env)
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	// Используем вызов без логгера, если он не требуется функцией
	dbPool, err := database.NewPostgresPool(context.Background(), dbDSN, log)
	if err != nil {
		log.Error("CRITICAL: Failed to initialize main database connection", slog.String("error", err.Error()))
		panic(fmt.Sprintf("failed to initialize database connection: %v", err))
	}
	log.Info("Successfully connected to the main database")

	// --- 2. Применение миграций БД ---
	// Получаем путь из ENV. Если не задан, используем путь по умолчанию для ЛОКАЛЬНОГО запуска.
	// Для Docker путь будет переопределен переменной окружения MIGRATIONS_PATH в docker-compose.yml
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "file://migrations" // Локальный путь по умолчанию (относительно корня проекта)
		log.Warn("MIGRATIONS_PATH environment variable not set, using default local path", slog.String("path", migrationsPath))
	} else {
		log.Info("Using migrations path from MIGRATIONS_PATH env var", slog.String("path", migrationsPath))
	}

	// Формируем URL для golang-migrate с схемой 'postgresql://'
	migrationDatabaseURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s&x-migrations-table=schema_migrations",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host, // Для Docker будет 'db', для локального - 'localhost'
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	log.Info("Attempting to apply database migrations...",
		slog.String("path", migrationsPath),
		slog.String("db", fmt.Sprintf("postgresql://%s:***@%s:%s/%s?sslmode=%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode)),
	)

	// Выполняем миграции с ретраями
	runMigrations(log, migrationsPath, migrationDatabaseURL)

	// --- 3. Инициализация остальных компонентов приложения ---
	return NewApp(context.Background(), cfg, log, dbPool)
}

// NewApp основной конструктор приложения.
// Инициализирует все компоненты: метрики, репозитории, сервисы, хендлеры, серверы.
// Принимает готовый пул соединений с БД.
func NewApp(ctx context.Context, cfg *config.Config, log *slog.Logger, dbPool *pgxpool.Pool) *App {
	const op = "app.NewApp"
	log = log.With(slog.String("op", op))
	log.Info("Initializing application components...")

	// Инициализация Метрик
	metricsCollector := promMetrics.NewCollector()
	metricsServer := promMetrics.RunMetricsServer(":" + cfg.Metrics.Port)
	log.Info("Metrics server configured", slog.String("port", cfg.Metrics.Port))

	// Инициализация Репозиториев
	pvzRepo := postgres.NewPVZRepository(dbPool, log)
	receptionRepo := postgres.NewReceptionRepository(dbPool, log)
	productRepo := postgres.NewProductRepository(dbPool, log)
	userRepo := postgres.NewUserRepository(dbPool, log)

	// Инициализация Хэшера
	hasher := hash.NewBcryptHasher(cfg.Hasher.BcryptCost)

	// Инициализация Сервисов
	authService := service.NewAuthService(log, cfg.Auth.JWTSecret, cfg.Auth.JWTttl, userRepo, hasher)
	pvzService := service.NewPVZService(log, pvzRepo, receptionRepo, metricsCollector)
	receptionService := service.NewReceptionService(log, pvzRepo, receptionRepo, metricsCollector)
	productService := service.NewProductService(log, receptionRepo, productRepo, metricsCollector)

	// Инициализация HTTP Хендлеров
	authHandler := httpHandler.NewAuthHandler(log, authService)
	pvzHandler := httpHandler.NewPVZHandler(log, pvzService, receptionService, productService)
	receptionHandler := httpHandler.NewReceptionHandler(log, receptionService)
	productHandler := httpHandler.NewProductHandler(log, productService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(mw.Recovery(log))
	logMiddleware := mw.NewLoggingMiddleware(log)
	router.Use(logMiddleware.LogRequest)
	router.Use(mw.PrometheusMiddleware(metricsCollector))
	authMiddleware := mw.NewAuthMiddleware(log, cfg.Auth.JWTSecret)

	router.POST("/dummyLogin", authHandler.PostDummyLogin)
	router.POST("/register", authHandler.PostRegister)
	router.POST("/login", authHandler.PostLogin)

	apiGroup := router.Group("/")
	apiGroup.Use(authMiddleware.Authorize)
	{
		pvzGroup := apiGroup.Group("/pvz")
		{
			pvzGroup.POST("", mw.RequireRole(domain.RoleModerator), pvzHandler.PostPvz)
			pvzGroup.GET("", mw.RequireRole(domain.RoleModerator, domain.RoleEmployee), pvzHandler.GetPvz)
			pvzWithIDGroup := pvzGroup.Group("/:pvzId")
			pvzWithIDGroup.Use(mw.RequireRole(domain.RoleEmployee))
			{
				pvzWithIDGroup.POST("/close_last_reception", pvzHandler.CloseLastReception)
				pvzWithIDGroup.POST("/delete_last_product", pvzHandler.DeleteLastProduct)
			}
		}
		receptionsGroup := apiGroup.Group("/receptions")
		receptionsGroup.Use(mw.RequireRole(domain.RoleEmployee))
		{
			receptionsGroup.POST("", receptionHandler.PostReceptions)
		}
		productsGroup := apiGroup.Group("/products")
		productsGroup.Use(mw.RequireRole(domain.RoleEmployee))
		{
			productsGroup.POST("", productHandler.PostProducts)
		}
	}

	// Инициализация gRPC Сервера
	grpcServer := grpcTransport.NewServer(log, pvzRepo, cfg.GRPCServer.Port)
	log.Info("gRPC server configured", slog.String("port", cfg.GRPCServer.Port))

	// Создание основного HTTP Сервера
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
	log.Info("HTTP server configured", slog.String("port", cfg.HTTPServer.Port))

	log.Info("Application components initialized successfully")

	return &App{
		cfg:           cfg,
		log:           log,
		dbPool:        dbPool,
		router:        router,
		server:        httpServer,
		grpcServer:    grpcServer,
		metricsServer: metricsServer,
	}
}

// runMigrations выполняет применение миграций БД с ретраями.
func runMigrations(log *slog.Logger, migrationsPath, databaseURL string) {
	const op = "app.runMigrations"
	log = log.With(slog.String("op", op))

	var m *migrate.Migrate
	var migrateErr error
	maxRetries := 5
	retryDelay := 3 * time.Second

	log.Info("Starting database migration process...")

	for attempt := 1; attempt <= maxRetries; attempt++ {
		m, migrateErr = migrate.New(migrationsPath, databaseURL)
		if migrateErr != nil {
			log.Warn("Failed to initialize migrate instance, retrying...",
				slog.Int("attempt", attempt), slog.Int("max_attempts", maxRetries),
				slog.String("error", migrateErr.Error()), slog.Duration("delay", retryDelay),
			)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
			}
			continue
		}

		log.Info("Applying migrations...", slog.Int("attempt", attempt))
		migrateErr = m.Up()
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Error("Error closing migrate source connection", slog.String("error", sourceErr.Error()))
		}
		if dbErr != nil {
			log.Error("Error closing migrate database connection", slog.String("error", dbErr.Error()))
		}

		if migrateErr == nil {
			log.Info("Database migrations applied successfully.")
			return // Успех
		}
		if errors.Is(migrateErr, migrate.ErrNoChange) {
			log.Info("Database schema is up to date. No changes applied.")
			return // Успех (нет изменений)
		}

		log.Warn("Failed to apply migrations, retrying...",
			slog.Int("attempt", attempt), slog.Int("max_attempts", maxRetries),
			slog.String("error", migrateErr.Error()), slog.Duration("delay", retryDelay),
		)
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	log.Error("FATAL: Could not apply database migrations after multiple attempts", slog.String("error", migrateErr.Error()))
	panic(fmt.Sprintf("could not apply database migrations: %v", migrateErr))
}

// GetRouter возвращает Gin Engine.
func (a *App) GetRouter() *gin.Engine {
	return a.router
}

// Run запускает все серверы приложения.
func (a *App) Run() {
	const op = "App.Run"
	log := a.log.With(slog.String("op", op))
	errChan := make(chan error, 3)

	go func() {
		log.Info("Starting main HTTP server", slog.String("address", a.server.Addr))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server ListenAndServe error", slog.String("error", err.Error()))
			errChan <- fmt.Errorf("http server failed: %w", err)
		} else {
			log.Info("HTTP server stopped listening")
		}
	}()

	go func() {
		if err := a.grpcServer.Start(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Error("gRPC server start error", slog.String("error", err.Error()))
			errChan <- fmt.Errorf("grpc server failed: %w", err)
		} else {
			log.Info("gRPC server stopped listening")
		}
	}()

	go func() {
		log.Info("Starting Metrics server", slog.String("address", a.metricsServer.Addr))
		if err := a.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Metrics server ListenAndServe error", slog.String("error", err.Error()))
			errChan <- fmt.Errorf("metrics server failed: %w", err)
		} else {
			log.Info("Metrics server stopped listening")
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Error("Server failed to start or run, initiating shutdown", slog.String("error", err.Error()))
	case sig := <-shutdownChan:
		log.Info("Shutdown signal received, initiating graceful shutdown", slog.String("signal", sig.String()))
	}

	log.Info("Starting graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	a.grpcServer.Stop()

	if err := a.metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Error("Metrics server graceful shutdown failed", slog.String("error", err.Error()))
	} else {
		log.Info("Metrics server stopped gracefully")
	}

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server graceful shutdown failed", slog.String("error", err.Error()))
	} else {
		log.Info("HTTP server stopped gracefully")
	}

	log.Info("Closing database connection pool...")
	a.dbPool.Close()
	log.Info("Database connection pool closed")

	log.Info("Graceful shutdown completed")
}
