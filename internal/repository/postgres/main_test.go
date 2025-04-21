package postgres_test

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/source/file"
	"log"
	"log/slog"
	"os"
	"pvz-service-avito-internship/internal/config"
	"pvz-service-avito-internship/internal/repository/postgres"
	"testing"

	"pvz-service-avito-internship/pkg/database"
	"pvz-service-avito-internship/pkg/logger"
)

var (
	dbPool     *pgxpool.Pool
	testConfig *config.Config
	testLogger *slog.Logger

	testUserRepo      *postgres.UserRepository
	testPVZRepo       *postgres.PVZRepository
	testReceptionRepo *postgres.ReceptionRepository
	testProductRepo   *postgres.ProductRepository
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Printf("WARN: .env file not found or error loading it: %v", err)
	}

	testConfig = config.LoadTestConfig()

	log.Printf("Loaded test DB config: host=%s, port=%s, user=%s, name=%s",
		testConfig.TestDatabase.Host, testConfig.TestDatabase.Port,
		testConfig.TestDatabase.User, testConfig.TestDatabase.Name)

	testDbDSN := database.BuildDSN(
		testConfig.TestDatabase.Host, testConfig.TestDatabase.Port, testConfig.TestDatabase.User,
		testConfig.TestDatabase.Password, testConfig.TestDatabase.Name, testConfig.TestDatabase.SSLMode,
	)
	testLogger = logger.Setup(testConfig.Logger.Level)

	var err error
	dbPool, err = database.NewPostgresPool(context.Background(), testDbDSN, testLogger)
	if err != nil {
		testLogger.Error("CRITICAL: Failed to connect to test database", slog.String("error", err.Error()))
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	testLogger.Info("Connected to TEST database successfully", slog.String("db_name", testConfig.TestDatabase.Name))
	testMigrationURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s&x-migrations-table=schema_migrations",
		testConfig.TestDatabase.User, testConfig.TestDatabase.Password, testConfig.TestDatabase.Host,
		testConfig.TestDatabase.Port, testConfig.TestDatabase.Name, testConfig.TestDatabase.SSLMode,
	)

	runTestMigrations(testLogger, "file://../../../migrations", testMigrationURL)

	testUserRepo = postgres.NewUserRepository(dbPool, testLogger)
	testPVZRepo = postgres.NewPVZRepository(dbPool, testLogger)
	testReceptionRepo = postgres.NewReceptionRepository(dbPool, testLogger)
	testProductRepo = postgres.NewProductRepository(dbPool, testLogger)

	exitCode := m.Run()

	testLogger.Info("Closing test database pool...")
	dbPool.Close()
	testLogger.Info("Test cleanup finished.")
	os.Exit(exitCode)
}

// runTestMigrations - применяет миграции к тестовой БД
func runTestMigrations(log *slog.Logger, migrationsPath, databaseURL string) {
	log = log.With(slog.String("op", "runTestMigrations"))
	log.Info("Attempting to apply migrations to TEST database...", slog.String("path", migrationsPath))
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		log.Error("FATAL: Failed to initialize migrate instance for test DB", slog.String("error", err.Error()))
		panic(fmt.Sprintf("Failed to init migrate for test DB: %v", err))
	}

	log.Info("Attempting to roll back any existing test migrations...")
	if errDown := m.Down(); errDown != nil && !errors.Is(errDown, migrate.ErrNoChange) && !errors.Is(errDown, os.ErrNotExist) /* ignore file not found for down */ {
		log.Warn("Failed to fully roll back test DB, proceeding anyway...", slog.String("error", errDown.Error()))
	} else if errDown == nil {
		log.Info("Previous test migrations rolled back.")
	}

	log.Info("Applying migrations up...")
	err = m.Up()
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		log.Error("Error closing migrate source connection", slog.String("error", sourceErr.Error()))
	}
	if dbErr != nil {
		log.Error("Error closing migrate db connection", slog.String("error", dbErr.Error()))
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("FATAL: Failed to apply migrations to test DB", slog.String("error", err.Error()))
		panic(fmt.Sprintf("Failed to apply migrations to test DB: %v", err))
	} else if errors.Is(err, migrate.ErrNoChange) {
		log.Info("Test database schema is up to date.")
	} else {
		log.Info("Test database migrations applied successfully.")
	}
}

// clearTables очищает указанные таблицы в тестовой БД
func clearTables(ctx context.Context, pool *pgxpool.Pool, tables ...string) error {
	log := testLogger.With(slog.String("op", "clearTables"))
	log.Info("Clearing test tables", slog.Any("tables", tables))
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for clearing tables: %w", err)
	}
	defer tx.Rollback(ctx) // Откатываем, если commit не удался

	for _, table := range tables {
		sql := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		log.Debug("Executing truncate", slog.String("sql", sql))
		if _, err := tx.Exec(ctx, sql); err != nil {
			log.Error("Failed to truncate table", slog.String("table", table), slog.String("error", err.Error()))
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction for clearing tables: %w", err)
	}
	log.Info("Test tables cleared successfully")
	return nil
}
