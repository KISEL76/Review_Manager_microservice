SERVICE_NAME=app
SERVICE_DIR=./microservice

UNIT_TESTS_DIR=./tests/unit/...
E2E_TESTS_DIR=./tests/e2e/...
E2E_NAME=review_manager_e2e

all: up

build:
	@docker-compose up --build -d 

up:
	@docker-compose up -d 

down:
	@docker-compose down

down-v:
	@docker-compose down -v 

logs:
	@docker-compose logs -f $(SERVICE_NAME)
	
test-unit:
	@echo "Запуск unit тестов..."
	@cd $(SERVICE_DIR) && go test -v $(UNIT_TESTS_DIR)
	@echo "Тестирование заверено"

test-e2e: 
	@echo "Запуск e2e теста с переменными окружения из .env.example..."
	@if [ -f .env ]; then cp .env .env.backup; fi
	@cp .env.example .env
	@COMPOSE_PROJECT_NAME=$(E2E_NAME) docker-compose up -d
	@echo "Тестирование по сценарию..."
	@cd $(SERVICE_DIR) && go test -v $(E2E_TESTS_DIR)	
	@cd $(SERVICE_DIR) && go test -v $(E2E_TESTS_DIR)	
	@COMPOSE_PROJECT_NAME=$(E2E_NAME) docker-compose down -v
	@if [ -f .env.backup ]; then mv .env.backup .env; else rm -f .env; fi
	@echo "Тестирование завершено"

coverage:
	@echo "Подсчет покрытия unit-тестов..."
	@cd $(SERVICE_DIR) && go test -v $(UNIT_TESTS_DIR) \
	  -coverpkg=review-manager/internal/service \
	  -coverprofile=coverage.out
	@cd $(SERVICE_DIR) && go tool cover -func=coverage.out
	@echo "Подсчет покрытия завершен"

coverage-html: coverage
	@echo "Открываю HTML-отчет покрытия..."
	@cd $(SERVICE_DIR) && go tool cover -html=coverage.out

fmt:
	@echo "Форматирование кода..."
	@goimports -w ./microservice/
	@echo "Форматирование завершено"

lint:
	@echo "Запуск линтера..."
	@cd microservice && golangci-lint run ./...
	@echo "Проверка линтером завершена"

clean:
	@echo "Очистка временных файлов..."
	@rm -f microservice/coverage.out
	@rm -f .env.backup
	@echo "Очистка завершена"

.PHONY: lint fmt test test-unit test-e2e coverage up down down-v logs build clean all