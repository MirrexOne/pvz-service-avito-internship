package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"pvz-service-avito-internship/internal/app"
	"pvz-service-avito-internship/internal/config"
	"pvz-service-avito-internship/pkg/database"
	"pvz-service-avito-internship/pkg/logger"
	"testing"

	"pvz-service-avito-internship/internal/handler/http/api"
)

var (
	testApp    *app.App
	testServer *httptest.Server
	dbPool     *pgxpool.Pool
	testConfig *config.Config
	testLogger *slog.Logger
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../.env"); err != nil {
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

	runTestMigrations(testLogger, "file://../migrations", testMigrationURL)
	testApp = app.NewApp(testConfig, testLogger, dbPool)
	testServer = httptest.NewServer(testApp.GetRouter())
	testLogger.Info("Test HTTP server started", slog.String("url", testServer.URL))
	exitCode := m.Run()
	testLogger.Info("Cleaning up test database...")
	clearTestDatabase(dbPool)
	testServer.Close()
	dbPool.Close()
	testLogger.Info("Test cleanup finished.")
	os.Exit(exitCode)
}

func runTestMigrations(log *slog.Logger, migrationsPath, databaseURL string) {
	log = log.With(slog.String("op", "RunTestMigrations"))
	log.Info("Attempting to apply migrations to TEST database...", slog.String("path", migrationsPath))

	m, err := migrate.New(migrationsPath, databaseURL)

	if err != nil {
		log.Error("FATAL: Failed to initialize migrate instance for test DB", slog.String("error", err.Error()))
		panic(fmt.Sprintf("Failed to init migrate for test DB: %v", err))
	}

	err = m.Up()
	_, dbErr := m.Close()

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

func clearTestDatabase(pool *pgxpool.Pool) {
	tables := []string{"products", "receptions", "pvz", "users"}

	for _, table := range tables {
		_, err := pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			testLogger.Error("Failed to truncate table", slog.String("table", table), slog.String("error", err.Error()))
		}
	}
	testLogger.Info("Test database tables truncated.")
}

func TestPVZFullFlow(t *testing.T) {
	clearTestDatabase(dbPool)
	require.NotNil(t, testServer, "Test server should be running")
	require.NotNil(t, testApp, "Test App should be initialized")
	client := testServer.Client()

	// Получение токенов
	moderatorToken := getDummyToken(t, client, api.UserRoleModerator)
	employeeToken := getDummyToken(t, client, api.UserRoleEmployee)

	// Создание ПВЗ
	createPVZReqBody := api.PVZ{City: api.Москва}
	var createdPVZ api.PVZ

	resp := doRequest(t, client, http.MethodPost, "/pvz", moderatorToken, createPVZReqBody, http.StatusCreated)
	err := json.Unmarshal(readBody(t, resp), &createdPVZ)
	require.NoError(t, err, "Failed to unmarshal created PVZ")
	require.NotNil(t, createdPVZ.Id, "Created PVZ ID should not be empty (*openapi_types.UUID)")
	assert.Equal(t, api.Москва, createdPVZ.City, "PVZ city mismatch")
	pvzID := *createdPVZ.Id
	t.Logf("Created PVZ ID: %s", pvzID)

	// Создание первого приема
	createRecReqBody := api.PostReceptionsJSONBody{PvzId: pvzID}
	var createdReception api.Reception

	resp = doRequest(t, client, http.MethodPost, "/receptions", employeeToken, createRecReqBody, http.StatusCreated)
	err = json.Unmarshal(readBody(t, resp), &createdReception)
	require.NoError(t, err, "Failed to unmarshal created reception")
	require.NotNil(t, createdReception.Id, "Created Reception ID should not be empty")
	assert.Equal(t, api.InProgress, createdReception.Status, "Reception should be in progress before adding products")
	t.Logf("Created new Reception ID: %s", *createdReception.Id)

	// Закрытие текущего приема
	closeRecURL := fmt.Sprintf("/pvz/%s/close_last_reception", pvzID.String())
	var closedReception api.Reception
	resp = doRequest(t, client, http.MethodPost, closeRecURL, employeeToken, nil, http.StatusOK)
	err = json.Unmarshal(readBody(t, resp), &closedReception)
	require.NoError(t, err, "Failed to unmarshal closed reception")
	require.NotNil(t, closedReception.Id, "Closed Reception ID should not be empty")
	assert.Equal(t, api.Close, closedReception.Status, "Reception status should be closed")
	t.Logf("Closed Reception ID: %s", closedReception.Id)

	// Создание нового приема
	resp = doRequest(t, client, http.MethodPost, "/receptions", employeeToken, createRecReqBody, http.StatusCreated)
	err = json.Unmarshal(readBody(t, resp), &createdReception)
	require.NoError(t, err, "Failed to unmarshal created reception")
	require.NotNil(t, createdReception.Id, "Created Reception ID should not be empty")
	t.Logf("Created new Reception ID: %s", *createdReception.Id)

	// Проверка статуса нового приема
	var receptionStatus api.ReceptionStatus
	err = dbPool.QueryRow(context.Background(), "SELECT status FROM receptions WHERE id = $1", createdReception.Id).Scan(&receptionStatus)
	require.NoError(t, err, "Failed to fetch reception status")
	require.Equal(t, api.InProgress, receptionStatus, "Reception should be in progress before adding products")

	t.Logf("Adding products to reception ID: %s", *createdReception.Id)

	// Добавление продуктов
	productType := api.ProductTypeОдежда
	t.Logf("Adding %d products...", 50)
	for i := 0; i < 50; i++ {
		addProductReqBody := api.PostProductsJSONBody{
			PvzId: pvzID,
			Type:  api.PostProductsJSONBodyType(productType),
		}

		var addedProduct api.Product
		resp = doRequest(t, client, http.MethodPost, "/products", employeeToken, addProductReqBody, http.StatusCreated)
		err = json.Unmarshal(readBody(t, resp), &addedProduct)
		require.NoErrorf(t, err, "Failed to unmarshal added product on iteration %d", i)
		require.NotNilf(t, addedProduct.Id, "Added product ID should not be empty on iteration %d", i)
		assert.Equalf(t, *createdReception.Id, addedProduct.ReceptionId, "Product reception ID mismatch on iteration %d", i)
		assert.Equalf(t, productType, addedProduct.Type, "Product type mismatch on iteration %d", i)
	}
	t.Log("Successfully added 50 products")

	// Закрытие приема после добавления продуктов
	resp = doRequest(t, client, http.MethodPost, closeRecURL, employeeToken, nil, http.StatusOK)
	err = json.Unmarshal(readBody(t, resp), &closedReception)
	require.NoError(t, err, "Failed to unmarshal closed reception")
	require.NotNil(t, closedReception.Id, "Closed Reception ID should not be empty")
	assert.Equal(t, api.Close, closedReception.Status, "Reception status should be closed")
	t.Logf("Closed Reception ID: %s", closedReception.Id)

	t.Log("PVZ Full Flow integration test completed successfully")
}

func TestPVZFullFlowWithError(t *testing.T) {
	clearTestDatabase(dbPool)
	require.NotNil(t, testServer, "Test server should be running")
	require.NotNil(t, testApp, "Test App should be initialized")
	client := testServer.Client()

	// Получение токенов
	moderatorToken := getDummyToken(t, client, api.UserRoleModerator)
	employeeToken := getDummyToken(t, client, api.UserRoleEmployee)

	// Создание ПВЗ
	createPVZReqBody := api.PVZ{City: api.Москва}
	var createdPVZ api.PVZ

	resp := doRequest(t, client, http.MethodPost, "/pvz", moderatorToken, createPVZReqBody, http.StatusCreated)
	err := json.Unmarshal(readBody(t, resp), &createdPVZ)
	require.NoError(t, err, "Failed to unmarshal created PVZ")
	require.NotNil(t, createdPVZ.Id, "Created PVZ ID should not be empty (*openapi_types.UUID)")
	assert.Equal(t, api.Москва, createdPVZ.City, "PVZ city mismatch")
	pvzID := *createdPVZ.Id
	t.Logf("Created PVZ ID: %s", pvzID)

	// Создание первого приема
	createRecReqBody := api.PostReceptionsJSONBody{PvzId: pvzID}
	var createdReception api.Reception

	resp = doRequest(t, client, http.MethodPost, "/receptions", employeeToken, createRecReqBody, http.StatusCreated)
	err = json.Unmarshal(readBody(t, resp), &createdReception)
	require.NoError(t, err, "Failed to unmarshal created reception")
	require.NotNil(t, createdReception.Id, "Created Reception ID should not be empty")
	assert.Equal(t, api.InProgress, createdReception.Status, "Reception should be in progress before adding products")
	t.Logf("Created new Reception ID: %s", *createdReception.Id)

	// Закрытие текущего приема
	closeRecURL := fmt.Sprintf("/pvz/%s/close_last_reception", pvzID.String())
	var closedReception api.Reception
	resp = doRequest(t, client, http.MethodPost, closeRecURL, employeeToken, nil, http.StatusOK)
	err = json.Unmarshal(readBody(t, resp), &closedReception)
	require.NoError(t, err, "Failed to unmarshal closed reception")
	require.NotNil(t, closedReception.Id, "Closed Reception ID) should not be empty")
	assert.Equal(t, api.Close, closedReception.Status, "Reception status should be closed")
	t.Logf("Closed Reception ID: %s", closedReception.Id)

	// Создание нового приема
	resp = doRequest(t, client, http.MethodPost, "/receptions", employeeToken, createRecReqBody, http.StatusCreated)
	err = json.Unmarshal(readBody(t, resp), &createdReception)
	require.NoError(t, err, "Failed to unmarshal created reception")
	require.NotNil(t, createdReception.Id, "Created Reception ID should not be empty")
	t.Logf("Created new Reception ID: %s", *createdReception.Id)

	// Проверка статуса нового приема
	var receptionStatus api.ReceptionStatus
	err = dbPool.QueryRow(context.Background(), "SELECT status FROM receptions WHERE id = $1", createdReception.Id).Scan(&receptionStatus)
	require.NoError(t, err, "Failed to fetch reception status")
	require.Equal(t, api.InProgress, receptionStatus, "Reception should be in progress before adding products")
	t.Logf("Adding products to reception ID: %s", *createdReception.Id)

	// Добавление продуктов
	productType := api.ProductTypeОдежда
	t.Logf("Adding %d products...", 50)
	for i := 0; i < 50; i++ {
		addProductReqBody := api.PostProductsJSONBody{
			PvzId: pvzID,
			Type:  api.PostProductsJSONBodyType(productType),
		}

		var addedProduct api.Product
		resp = doRequest(t, client, http.MethodPost, "/products", employeeToken, addProductReqBody, http.StatusCreated)
		err = json.Unmarshal(readBody(t, resp), &addedProduct)
		require.NoErrorf(t, err, "Failed to unmarshal added product on iteration %d", i)
		require.NotNilf(t, addedProduct.Id, "Added product ID should not be empty on iteration %d", i)
		assert.Equalf(t, *createdReception.Id, addedProduct.ReceptionId, "Product reception ID mismatch on iteration %d", i)
		assert.Equalf(t, productType, addedProduct.Type, "Product type mismatch on iteration %d", i)
	}

	t.Log("Successfully added 50 products")

	// Закрытие приема после добавления продуктов
	resp = doRequest(t, client, http.MethodPost, closeRecURL, employeeToken, nil, http.StatusOK)
	err = json.Unmarshal(readBody(t, resp), &closedReception)
	require.NoError(t, err, "Failed to unmarshal closed reception")
	require.NotNil(t, closedReception.Id, "Closed Reception ID should not be empty")
	assert.Equal(t, api.Close, closedReception.Status, "Reception status should be closed")
	t.Logf("Closed Reception ID: %s", closedReception.Id)

	// Попытка закрыть прием повторно
	resp = doRequest(t, client, http.MethodPost, closeRecURL, employeeToken, nil, http.StatusBadRequest)
	err = json.Unmarshal(readBody(t, resp), &closedReception)
	require.NoError(t, err, "Failed to unmarshal closed reception")
	require.NotNil(t, closedReception.Id, "Closed Reception ID should not be empty")
	assert.Equal(t, api.Close, closedReception.Status, "Reception status should be closed")
	t.Logf("Closed Reception ID: %s", closedReception.Id)

	// Попытка добавить продукт в закрытую приемку - должна вернуть ошибку
	addProductReqBody := api.PostProductsJSONBody{
		PvzId: pvzID,
		Type:  api.PostProductsJSONBodyType(productType),
	}

	resp = doRequest(t, client, http.MethodPost, "/products", employeeToken, addProductReqBody, http.StatusBadRequest)

	// Проверяем, что получили именно сообщение об ошибке, а не объект продукта
	var errorResp map[string]interface{}
	err = json.Unmarshal(readBody(t, resp), &errorResp)
	require.NoError(t, err, "Не удалось распарсить ответ с ошибкой")
	require.Contains(t, errorResp, "message", "Ответ должен содержать поле 'message'")
	t.Logf("Получена ожидаемая ошибка при попытке добавить продукт в закрытую приемку: %v", errorResp["message"])
}

func getDummyToken(t *testing.T, client *http.Client, role api.UserRole) string {
	t.Helper()

	reqBody := api.PostDummyLoginJSONBody{Role: api.PostDummyLoginJSONBodyRole(role)}
	resp := doRequest(t, client, http.MethodPost, "/dummyLogin", "", reqBody, http.StatusOK)

	var tokenResp api.Token
	bodyBytes := readBody(t, resp)
	err := json.Unmarshal(bodyBytes, &tokenResp)
	require.NoErrorf(t, err, "Failed to unmarshal token response for role %s. Body: %s", role, string(bodyBytes))
	require.NotEmptyf(t, tokenResp, "Token should not be empty for role %s", role)
	return tokenResp
}

func doRequest(t *testing.T, client *http.Client, method, path, token string, body interface{}, expectedStatus int) *http.Response {
	t.Helper()
	var reqBodyReader io.Reader = nil
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoErrorf(t, err, "Failed to marshal request body for %s %s", method, path)
		reqBodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, testServer.URL+path, reqBodyReader)
	require.NoErrorf(t, err, "Failed to create request for %s %s", method, path)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	require.NoErrorf(t, err, "Failed to execute request for %s %s", method, path)
	bodyBytes := readBody(t, resp)
	require.Equalf(t, expectedStatus, resp.StatusCode, "Unexpected status code for %s %s. Response body: %s", method, path, string(bodyBytes))
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return resp
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoErrorf(t, err, "Failed to read response body")
	return body
}
