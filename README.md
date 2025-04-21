# Сервис для работы с ПВЗ (Avito Backend Internship - Весна 2025)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://golang.org/dl/)
[![Docker](https://img.shields.io/badge/Docker%20&%20Compose-Required-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Required-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org/)

Бэкенд-сервис на Go, разработанный в рамках тестового задания для стажировки в Avito. Сервис предназначен для
сотрудников Пунктов Выдачи Заказов (ПВЗ) и позволяет вносить информацию о приемках товаров, управлять товарами в рамках
приемки и получать отчетность по ПВЗ с фильтрацией и пагинацией.

## Основной Функционал

* **Аутентификация/Авторизация:**
    * Регистрация (`/register`) и вход (`/login`) пользователей с ролями `employee` (сотрудник ПВЗ) и `moderator` по
      email/паролю.
    * Генерация JWT токенов для доступа к защищенным ресурсам.
    * Тестовый эндпоинт `/dummyLogin` для быстрого получения токена с нужной ролью.
    * Защита эндпоинтов с использованием JWT и проверки ролей.
* **Управление ПВЗ:**
    * Создание ПВЗ (`POST /pvz`) только модератором в разрешенных городах (Москва, Санкт-Петербург, Казань).
    * Получение списка ПВЗ (`GET /pvz`) с пагинацией и фильтрацией по дате приемки товаров (доступно модератору и
      сотруднику). Ответ включает детальную информацию о приемках и товарах в этих ПВЗ.
* **Управление Приемками Товаров:**
    * Создание новой приемки (`POST /receptions`) сотрудником для конкретного ПВЗ. Невозможно создать, если есть
      предыдущая незакрытая приемка.
    * Закрытие последней активной приемки (`POST /pvz/{pvzId}/close_last_reception`) сотрудником.
* **Управление Товарами в Приемке:**
    * Добавление товара (`POST /products`) сотрудником в текущую активную приемку ПВЗ. Поддерживаются типы:
      `электроника`, `одежда`, `обувь`.
    * Удаление последнего добавленного товара (`POST /pvz/{pvzId}/delete_last_product`) сотрудником из активной приемки
      по принципу LIFO.

## Дополнительные Возможности (Реализованы)

* **gRPC Интерфейс:** Метод `GetPVZList` для получения полного списка всех ПВЗ без авторизации (порт 3000).
* **Мониторинг Prometheus:** Сбор технических (HTTP запросы, время ответа) и бизнесовых метрик (созданные ПВЗ, приемки,
  товары). Метрики доступны по эндпоинту `/metrics` (порт 9000).
* **Структурированное Логирование:** Используется `slog` с JSON-форматом и Request ID для трассировки.
* **Автоматические Миграции БД:** Схема PostgreSQL создается и обновляется автоматически при старте приложения с помощью
  `golang-migrate/migrate`.
* **Кодогенерация:**
    * DTO и серверные интерфейсы для HTTP API генерируются из `api/swagger.yaml` с помощью `oapi-codegen`.
    * gRPC код генерируется из `api/pvz.proto` с помощью `protoc`.
* **Тестирование:** Проект содержит unit и интеграционные тесты (требуется запуск для проверки покрытия).
* **Архитектура:** Приложение построено с использованием слоистой архитектуры (Handler, Service, Repository) для лучшей
  тестируемости и поддержки.

## Используемые Технологии

* **Язык:** Go (1.24)
* **Веб-фреймворк:** Gin
* **База данных:** PostgreSQL
* **Драйвер БД:** pgx/v5
* **Query Builder:** Squirrel
* **Миграции:** golang-migrate/migrate (применяются автоматически при старте приложения)
* **Авторизация:** JWT (golang-jwt/jwt/v5), bcrypt
* **Конфигурация:** cleanenv (ENV + опционально YAML), godotenv (для локального `.env`)
* **Логирование:** slog (стандартная библиотека, JSON формат)
* **Кодогенерация:**
    * oapi-codegen (HTTP DTO из OpenAPI)
    * protoc + Go gRPC плагины (для gRPC)
    * mockery (для генерации моков)
* **Мониторинг:** Prometheus (client_golang)
* **Тестирование:** testify (assert, require, mock), net/http/httptest, bufconn (для gRPC тестов)
* **Контейнеризация:** Docker, Docker Compose

## Подготовка к Запуску

1. **Установите Go:** Версии 1.24.
2. **Установите Docker и Docker Compose.**
3. **Установите инструменты для генерации кода:**
    * `protoc`: [Инструкция](https://grpc.io/docs/protoc-installation/) (Скачайте архив для вашей ОС, распакуйте,
      добавьте путь к `bin` в PATH).
    * Go gRPC плагины:
      ```bash
      go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
      go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
      ```
    * `oapi-codegen`:
      ```bash
      go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
      ```
    * `mockery` (для генерации моков):
      ```bash
      go install github.com/vektra/mockery/v2@latest
      ```
    * `golang-migrate` CLI (опционально, для ручных миграций тестовой
      БД): [Инструкция](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation)
    * **ВАЖНО:** Убедитесь, что директории установки (`$GOPATH/bin` или `$GOBIN`, путь к `protoc/bin`) добавлены в
      системную переменную `PATH`.
4. **Клонируйте репозиторий:**
   ```bash
   git clone https://github.com/MirrexOne/pvz-service-avito-internship.git
   cd pvz-service-avito-internship
   ```
5. **Создайте `.env` файл:** Скопируйте `.env.example` в `.env`. **ОБЯЗАТЕЛЬНО** измените `JWT_SECRET` на надежный ключ.
   Настройте параметры `DB_*` для основной БД и `TEST_DB_*` для тестовой БД.
   ```bash
   cp .env.example .env
   # Откройте .env и отредактируйте значения
   ```
6. **Сгенерируйте код:** Выполните скрипт из корневой директории проекта:
   ```bash
   ./scripts/generate.sh
   ```
   *(Эта команда запустит `oapi-codegen`, `protoc` и `mockery`)*
7. **Скачайте/Обновите зависимости Go:**
   ```bash
   go mod tidy
   ```

## Запуск с Docker Compose (Рекомендуемый)

Этот способ запускает приложение и основную базу данных PostgreSQL в контейнерах. Миграции для основной БД применяются автоматически при старте контейнера приложения.

1.  **Установить Docker и Docker Compose:** Если они еще не установлены на целевой машине. [Инструкция Docker](https://docs.docker.com/engine/install/), [Инструкция Compose](https://docs.docker.com/compose/install/).
2.  **Получить Код Проекта:**
    *   Либо клонировать репозиторий: `git clone <your-repo-url>`
    *   Либо скопировать всю папку с проектом на целевую машину.
3.  **Перейти в Корневую Директорию Проекта:** `cd path/to/pvz-service-avito-internship`
4.  **Создать и Настроить `.env` Файл:**
    *   Скопировать `.env.example` в `.env`: `cp .env.example .env`
    *   **Обязательно** установить уникальный `JWT_SECRET`.
    *   Проверить и при необходимости **адаптировать** переменные `DB_*` (особенно `DB_PASSWORD`, если используется другой), `*_PORT` (если порты на хосте заняты). **`DB_HOST` должен остаться `db`**, так как это имя сервиса внутри Docker Compose.
    *   Переменные `TEST_DB_*` не обязательны, если на этой машине не будут запускаться интеграционные тесты.
5.  **Запустить Docker Compose:**
    ```bash
    docker-compose up --build -d
    ```
    *   `--build`: Важен при первом запуске на новой машине, чтобы собрать образы.
    *   `-d`: Запустить в фоновом режиме.

**Дополнительные зависимости (Go, protoc, etc.) на целевой машине НЕ ТРЕБУЮТСЯ** для простого запуска через `docker-compose`, так как все необходимое для сборки и запуска приложения уже описано в `Dockerfile` и будет выполнено внутри контейнеров.

## Локальный Запуск (Для Отладки)

Этот способ требует вручную запустить PostgreSQL и установить переменные окружения.

1. **Запустите PostgreSQL:** Локально или через `docker-compose up db`. Убедитесь, что база (`DB_NAME`) и пользователь (
   `DB_USER`) существуют.
2. **Установите переменные окружения:** Задайте переменные из `.env` файла для вашего терминала или конфигурации запуска
   IDE. **Критически важны:**
    * `DB_HOST=localhost` (или адрес вашего локального Postgres)
    * `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
    * `JWT_SECRET`
    * `MIGRATIONS_PATH=file://migrations` (путь к локальной папке миграций)
3. **Запустите приложение из корня проекта:**
   ```bash
   go run ./cmd/server/main.go
   ```
   *(Миграции для основной БД применятся автоматически)*

## Остановка Docker Compose

```bash
docker-compose down        # Остановить и удалить контейнеры
docker-compose down -v     # Остановить, удалить контейнеры и тома (данные БД!)
```

## API и Документация

- OpenAPI (Swagger) Спецификация: Файл api/swagger.yaml. Импортируйте в Postman или откройте в Swagger Editor для
  интерактивной документации.

- gRPC Спецификация: Файл api/pvz.proto.

- Примеры Запросов Postman/Curl: См. раздел Examples ниже или используйте спецификацию.

## Запуск Тестов

1. Настройте тестовую БД:

    - Убедитесь, что у вас запущен PostgreSQL сервер, доступный по адресу и порту из переменных TEST_DB_HOST и
      TEST_DB_PORT_HOST (например, localhost:5433).
    - Убедитесь, что база данных (TEST_DB_NAME), пользователь (TEST_DB_USER) и пароль (TEST_DB_PASSWORD) из .env
      существуют
      и пользователь имеет права на создание/удаление таблиц в тестовой БД.

2. Примените миграции к тестовой БД вручную:

```bash
  # Замените переменные на значения TEST_DB_* из .env

export TEST_DATABASE_URL="postgresql://${TEST_DB_USER}:${TEST_DB_PASSWORD}@${TEST_DB_HOST}:${TEST_DB_PORT_HOST}/${TEST_DB_NAME}?sslmode=${TEST_DB_SSL_MODE}"
migrate -path ./migrations -database $TEST_DATABASE_URL up
```

3. Сгенерируйте моки (если требуется обновить):

```bash
  mockery --config=mockery.yaml
```

4. Запустите все тесты из корня проекта:

```bash
  go test -race -cover ./...
```

5. Просмотр покрытия:

```bash
  go test -race -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out
```

## Examples (Примеры Запросов Curl) <a name="examples"></a>

- [Регистрация](#register-login)
- [Аутентификация](#register-login)
- [Создание ПВЗ](#create-pvz)
- [Создание Приемки](#create-reception)
- [Добавление Товара](#add-product)
- [Удаление Товара](#delete-product)
- [Закрытие Приемки](#close-reception)
- [Получение Списка ПВЗ](#list-pvz)
- [Получение Списка ПВЗ (gRPC)](#grpc-list-pvz)
- [Получение Метрик Prometheus](#prometheus-metrics)

### Получение токена (Dummy) <a name="dummy-login"></a>

#### Получить токен модератора

```curl
curl --location --request POST 'http://localhost:8080/dummyLogin' \
--header 'Content-Type: application/json' \
--data-raw '{ "role": "moderator" }'
```

#### Получить токен сотрудника

```curl
curl --location --request POST 'http://localhost:8080/dummyLogin' \
--header 'Content-Type: application/json' \
--data-raw '{ "role": "employee" }'
```

Пример ответа:

```json
"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Регистрация и Вход <a name="register-login"></a>

#### Регистрация сотрудника

```curl
curl --location --request POST 'http://localhost:8080/register' \
--header 'Content-Type: application/json' \
--data-raw '{
"email": "new_employee@example.com",
"password": "password123",
"role": "employee"
}'
```
Пример ответа:

```json
{
"id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
"email": "new_employee@example.com",
"role": "employee"
}
```

#### Вход пользователя

```curl
curl --location --request POST 'http://localhost:8080/login' \
--header 'Content-Type: application/json' \
--data-raw '{
"email": "new_employee@example.com",
"password": "password123"
}'
```
Пример ответа:
```json
"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3..."
```

### Создание ПВЗ <a name="create-pvz"></a>

```curl
curl -X POST 'http://localhost:8080/pvz' \
-H 'Authorization: Bearer <YOUR_MODERATOR_TOKEN>' \
-H 'Content-Type: application/json' \
-d '{
    "city": "Казань"
}'
```

Пример ответа:

```json
{
  "id": "a1b2c3d4-e5f6-7890-1234-abcdef123456",
  "registrationDate": "2025-04-19T12:00:00Z",
  "city": "Казань"
}
```

### Создание Приемки <a name="create-reception"></a>

Замените YOUR_PVZ_ID на ID, полученный при создании ПВЗ

```curl
curl -X POST 'http://localhost:8080/receptions' \
-H 'Authorization: Bearer <YOUR_EMPLOYEE_TOKEN>' \
-H 'Content-Type: application/json' \
-d '{
    "pvzId": "<YOUR_PVZ_ID>"
}'
```

Пример ответа:

```json
{
  "id": "e5f6g7h8-...",
  "dateTime": "2025-04-19T12:05:00Z",
  "pvzId": "<YOUR_PVZ_ID>",
  "status": "in_progress"
}
```

### Добавление Товара <a name="add-product"></a>

```curl
curl -X POST 'http://localhost:8080/products' \
-H 'Authorization: Bearer <YOUR_EMPLOYEE_TOKEN>' \
-H 'Content-Type: application/json' \
-d '{
    "pvzId": "<YOUR_PVZ_ID>",
    "type": "одежда"
}'
```

Пример ответа:

```json
{
  "id": "i9j0k1l2-...",
  "dateTime": "2025-04-19T12:06:00Z",
  "type": "одежда",
  "receptionId": "<YOUR_RECEPTION_ID>"
}
```

### Удаление Товара <a name="delete-product"></a>

```curl
curl -X POST 'http://localhost:8080/pvz/<YOUR_PVZ_ID>/delete_last_product' \
-H 'Authorization: Bearer <YOUR_EMPLOYEE_TOKEN>'
```

Пример ответа:

```text
Статус 200 OK (без тела)
```

### Закрытие Приемки <a name="close-reception"></a>

```curl
curl -X POST 'http://localhost:8080/pvz/<YOUR_PVZ_ID>/close_last_reception' \
-H 'Authorization: Bearer <YOUR_EMPLOYEE_TOKEN>'
```

Пример ответа:

```json
{
  "id": "<YOUR_RECEPTION_ID>",
  "dateTime": "2025-04-19T12:05:00Z",
  "pvzId": "<YOUR_PVZ_ID>",
  "status": "close"
}
```

### Получение Списка ПВЗ <a name="list-pvz"></a>

Пример: первая страница, 5 элементов, фильтр по дате

```curl
curl -X GET 'http://localhost:8080/pvz?page=1&limit=2&startDate=2025-04-19T00:00:00Z' \
-H 'Authorization: Bearer <YOUR_EMPLOYEE_TOKEN_OR_MODERATOR_TOKEN>'
```

Пример ответа:

```json
{
  "items": [
    {
      "pvz": {
        "id": "a1b2c3d4-...",
        "registrationDate": "2025-04-18T10:00:00Z",
        "city": "Москва"
      },
      "receptions": [
        {
          "reception": {
            "id": "e5f6g7h8-...",
            "dateTime": "2025-04-19T12:05:00Z",
            "pvzId": "a1b2c3d4-...",
            "status": "close"
          },
          "products": [
            {
              "id": "m3n4o5p6-...",
              "dateTime": "2025-04-19T12:07:00Z",
              "type": "электроника",
              "receptionId": "e5f6g7h8-..."
            }
          ]
        }
      ]
    }
    // ... другие ПВЗ
  ],
  "total": 1,
  "page": 1,
  "limit": 5
}
```
### Получение Списка ПВЗ (gRPC) <a name="grpc-list-pvz"></a>
(Требуется установленный grpcurl или можно использовать gRPC клиент Postman)
```bash 
  grpcurl -plaintext localhost:3000 pvz.v1.PVZService/GetPVZList
```
Пример ответа:
```json
{
  "pvz": [
    {
      "id": "a1b2c3d4-...",
      "registrationDate": "2025-04-18T10:00:00Z",
      "city": "Москва"
    },
    {
      "id": "e5f6g7h8-...",
      "registrationDate": "2025-04-19T12:00:00Z",
      "city": "Казань"
    }
  ]
}
```
### Получение Метрик Prometheus <a name="prometheus-metrics"></a>
```curl
curl http://localhost:9000/metrics
# Или просто откройте http://localhost:9000/metrics в браузере
```

Пример ответа:
```text
# HELP pvz_created_total Total number of PVZs created.
# TYPE pvz_created_total counter
pvz_created_total 3
# HELP pvz_http_request_duration_seconds Histogram of HTTP request durations in seconds.
# TYPE pvz_http_request_duration_seconds histogram
pvz_http_request_duration_seconds_bucket{method="POST",path="/login",le="0.005"} 1
pvz_http_request_duration_seconds_bucket{method="POST",path="/login",le="0.01"} 1
...
# HELP pvz_http_requests_total Total number of processed HTTP requests.
# TYPE pvz_http_requests_total counter
pvz_http_requests_total{method="GET",path="/pvz",status_code="200"} 5
pvz_http_requests_total{method="POST",path="/login",status_code="200"} 1
pvz_http_requests_total{method="POST",path="/pvz",status_code="201"} 3
...
```

### Принятые Решения <a name="decisions"></a>

1. Автоматические миграции:

> Миграции БД применяются автоматически при старте Go-приложения для удобства запуска.
Используется golang-migrate/migrate. Путь к файлам миграций определяется переменной окружения MIGRATIONS_PATH (для
Docker устанавливается в file:///app/migrations, для локального запуска по умолчанию file://migrations).

2. LIFO Удаление Товаров:

> Реализовано через поиск последнего добавленного товара по полю date_time в таблице products (ORDER BY date_time DESC
LIMIT 1).

3. Конфигурация:

> Приоритет отдается переменным окружения (читаются cleanenv). Файл .env используется для задания переменных
при запуске через Docker Compose или локально (через godotenv в config.Load). Файл configs/config.yaml может содержать
дефолтные значения (не переменные ${...}), но будет переопределен ENV.

4. Тестовая БД: 
> Интеграционные тесты (go test ./integration/...) требуют отдельной, вручную подготовленной (БД,
пользователь, права) и мигрированной (через migrate CLI) тестовой базы данных. Параметры задаются через TEST_DB_* в
.env. TestMain в тестах репозитория (internal/repository/postgres/) также автоматически применяет миграции и очищает
тестовую БД для этих тестов.

5. Образ Приложения: 
> Используется alpine в качестве финального Docker-образа для возможности запуска миграций из Go кода (
требует libpq).

6. Зависимости Тестов: 
> Unit-тесты используют моки (mockery). Интеграционные тесты репозиториев и сквозные тесты работают с
реальной тестовой БД.