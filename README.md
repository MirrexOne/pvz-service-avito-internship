# pvz-service-avito-internship

# Сервис для работы с ПВЗ

Backend-сервис на Go для управления Пунктами Выдачи Заказов (ПВЗ), приемками товаров и товарами в рамках этих приемок.
Реализует REST API, gRPC интерфейс, сбор метрик Prometheus и автоматические миграции базы данных.

[![codecov](https://codecov.io/gh/MirrexOne/pvz-service-avito-internship/graph/badge.svg?token=9V692WYGM3)](https://codecov.io/gh/MirrexOne/pvz-service-avito-internship)

## Функционал

* **Аутентификация/Авторизация:**
    * Регистрация пользователей (`/register`) с ролями `employee` и `moderator`.
    * Вход пользователей (`/login`) по email/паролю с выдачей JWT токена.
    * Dummy-вход (`/dummyLogin`) для получения тестового токена с нужной ролью.
    * Защита эндпоинтов с использованием JWT и проверки ролей.
* **Управление ПВЗ:**
    * Создание ПВЗ (`POST /pvz`) модератором в разрешенных городах (Москва, Санкт-Петербург, Казань).
    * Получение списка ПВЗ (`GET /pvz`) с пагинацией и фильтрацией по дате приемки товаров.
* **Управление Приемками:**
    * Создание новой приемки (`POST /receptions`) сотрудником. Блокируется, если есть незакрытая.
    * Закрытие последней активной приемки (`POST /pvz/{pvzId}/close_last_reception`) сотрудником.
* **Управление Товарами:**
    * Добавление товара (`POST /products`) сотрудником в текущую активную приемку.
    * Удаление последнего добавленного товара (`POST /pvz/{pvzId}/delete_last_product`) сотрудником из активной
      приемки (LIFO).
* **gRPC:** Метод `GetPVZList` для получения полного списка ПВЗ без авторизации.
* **Мониторинг:** Сбор метрик Prometheus (технические и бизнесовые), эндпоинт `/metrics`.
* **Логирование:** Структурированное `slog` (JSON), Request ID.
* **База Данных:** PostgreSQL, автоматические миграции при старте приложения (`golang-migrate/migrate`).
* **Технологии:** Go, Gin, pgx/v5, Squirrel, JWT, gRPC, Prometheus, Docker, oapi-codegen, mockery.

## Структура Проекта

(Структура папок, как описано ранее)

## Подготовка к Запуску

1. **Установите Go:** Версии 1.22 или выше.
2. **Установите Docker и Docker Compose.**
3. **Установите инструменты для генерации кода:**
    * `protoc`: [Инструкция](https://grpc.io/docs/protoc-installation/)
    * Go gRPC плагины:
      `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`
    * `oapi-codegen`: `go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest`
    * `mockery`: `go install github.com/vektra/mockery/v2@latest`
    * `golang-migrate` CLI (для ручных
      миграций/тестов): [Инструкция](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation)
    * **ВАЖНО:** Убедитесь, что пути к этим инструментам добавлены в ваш системный `PATH`.
4. **Клонируйте репозиторий.**
5. **Создайте `.env` файл:** Скопируйте `.env.example` в `.env`. **ОБЯЗАТЕЛЬНО** измените `JWT_SECRET`. Настройте
   параметры `DB_*` и `TEST_DB_*`.
6. **Сгенерируйте код:**
   ```bash
   ./scripts/generate.sh && mockery --config=mockery.yaml
   ```
7. **Скачайте зависимости Go:** `go mod tidy`

## Запуск с Docker Compose (Рекомендуемый)

1. **Убедитесь, что Docker Desktop запущен.**
2. **Запустите из корневой директории проекта:**
   ```bash
   docker-compose up --build -d
   ```
   Приложение запустится, и миграции для основной БД будут применены автоматически при старте контейнера `app`.
3. **Сервис будет доступен:**
    * HTTP API: `http://localhost:8080` (порт из `.env`)
    * gRPC API: `localhost:3000` (порт из `.env`)
    * Prometheus Metrics: `http://localhost:9000/metrics` (порт из `.env`)

## Локальный Запуск (Для Отладки)

1. **Запустите PostgreSQL:** Локально или через `docker-compose up db` (только БД). Убедитесь, что база данных (
   `DB_NAME`) и пользователь (`DB_USER`) существуют.
2. **Установите переменные окружения:** Задайте переменные из `.env` файла для вашего терминала или конфигурации запуска
   IDE. **Критически важны:**
    * `DB_HOST=localhost` (или адрес вашего локального Postgres)
    * `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
    * `JWT_SECRET`
    * `MIGRATIONS_PATH=file://migrations` (путь к локальной папке миграций)
3. **Примените миграции вручную:**
   ```bash
   migrate -path ./migrations -database 'postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}' up
   ```
4. **Запустите приложение из корня проекта:**
   ```bash
   go run ./cmd/server/main.go
   ```

## Остановка Docker Compose

```bash
docker-compose down # Остановка контейнеров
docker-compose down -v # Остановка и удаление томов (данных БД)