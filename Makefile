.PHONY: run build test migrate clean

# Переменные
BINARY_NAME=cosmos-api
MIGRATIONS_DIR=migrations

# Запуск
run:
	@echo "Starting server..."
	go run cmd/api/main.go

# Сборка
build:
	@echo "Building..."
	go build -o bin/$(BINARY_NAME) cmd/api/main.go

# Запуск собранного бинарника
start: build
	@echo "Starting $(BINARY_NAME)..."
	./bin/$(BINARY_NAME)

# Тесты
test:
	@echo "Running tests..."
	go test ./... -v

# Миграции (требуется psql)
migrate:
	@echo "Applying migrations..."
	psql -U postgres -d cosmos -f $(MIGRATIONS_DIR)/001_init.sql

# Очистка
clean:
	@echo "Cleaning..."
	go clean
	rm -rf bin/
	rm -f .env

# Установка зависимостей
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Форматирование кода
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Проверка кода
lint:
	@echo "Linting..."
	golangci-lint run

# Полный запуск проекта с миграциями
setup: deps migrate run

help:
	@echo "Available commands:"
	@echo "  make run     - запустить приложение"
	@echo "  make build   - собрать бинарник"
	@echo "  make start   - собрать и запустить"
	@echo "  make test    - запустить тесты"
	@echo "  make migrate - применить миграции"
	@echo "  make clean   - очистить проект"
	@echo "  make deps    - установить зависимости"
	@echo "  make fmt     - отформатировать код"
	@echo "  make setup   - полная установка и запуск"
