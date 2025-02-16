include .env
export ${shell sed 's/=.*//' .env}

PYTHON_VERSION = python3.10
BINARY_NAME = avito_merch_store

.DEFAULT_GOAL := help

DB_SCRIPT_VENV_DIR = ./scripts/database/.venv
DB_SETUP_SCRIPT = ./scripts/database/setup.py
DB_SCRIPT_REQ_FILE = ./scripts/database/requirements.txt

TEST_DB_SETUP_SCRIPT = ./scripts/test_database/setup.py
TEST_DB_SCRIPT_REQ_FILE = ./scripts/test_database/requirements.txt

help:
	@echo "Makefile команды:"
	@echo "  all     - Полный цикл: tidy, fmt, lint, test, build"
	@echo "  build   - Собрать бинарный файл"
	@echo "  run     - Собрать и запустить приложение"
	@echo "  clean   - Удалить собранные файлы"
	@echo "  test    - Запустить тесты"
	@echo "  fmt     - Отформатировать код"
	@echo "  lint    - Запустить линтер (golangci-lint)"
	@echo "  tidy    - Обновить зависимости"

.PHONY: all build clean 

# --- Additional scripts ---

# Main Database
db-script-setup:
	@echo "Setting up Python virtual environment..."
	${PYTHON_VERSION} -m venv ${DB_SCRIPT_VENV_DIR}
	@echo "Activating virtual environment and installing dependencies..."
	${DB_SCRIPT_VENV_DIR}/bin/pip3.10 install --upgrade pip
	${DB_SCRIPT_VENV_DIR}/bin/pip3.10 install -r ${DB_SCRIPT_REQ_FILE}
	@echo "Running setup.py..."
	${DB_SCRIPT_VENV_DIR}/bin/python ${DB_SETUP_SCRIPT}

db-script-clean:
	rm -rf ${DB_SCRIPT_VENV_DIR}

# Test Database
test-db-script-create:
	bash ./scripts/test_database/create.sh

test-db-script-setup:
	@echo "Setting up Python virtual environment..."
	${PYTHON_VERSION} -m venv ./scripts/test_database/.venv
	@echo "Activating virtual environment and installing dependencies..."
	./scripts/test_database/.venv/bin/pip3.10 install --upgrade pip
	./scripts/test_database/.venv/bin/pip3.10 install -r ${TEST_DB_SCRIPT_REQ_FILE}
	@echo "Running setup.py..."
	./scripts/test_database/.venv/bin/${PYTHON_VERSION} ./scripts/test_database/setup.py

test-db-script-delete:
	bash ./scripts/test_database/delete.sh

test-db-script-clean: test-db-script-delete
	rm -rf ./scripts/test_database/.venv

#  Integration tests
test-int:
	docker compose -f ./scripts/integration_test/docker-compose-test.yaml up -d
	@echo "Setting up Python virtual environment..."
	${PYTHON_VERSION} -m venv ./scripts/integration_test/.venv
	@echo "Activating virtual environment and installing dependencies..."
	./scripts/integration_test/.venv/bin/pip3.10 install --upgrade pip
	./scripts/integration_test/.venv/bin/pip3.10 install -r ./scripts/integration_test/requirements.txt
	@echo "Running setup.py..."
	./scripts/integration_test/.venv/bin/${PYTHON_VERSION} ./scripts/integration_test/setup.py
	@echo "Integration tests starting"
	./scripts/integration_test/.venv/bin/${PYTHON_VERSION} ./scripts/integration_test/api_tests.py
	docker compose -f ./scripts/integration_test/docker-compose-test.yaml down


test-unit:
	go test -coverpkg=./api/controller -coverprofile=coverage.out ./test/unit/controller/

test-bench: ## run benchmark tests
	go test -bench ./...

# Generate test coverage
test-all: test-db-script-delete test-db-script-create test-db-script-setup test-unit test-int

# --- Main scripts ---

# all: tidy fmt lint test build
all: tidy clean build

build:
	mkdir -p ./.bin
	go build -o ./.bin/${BINARY_NAME} ./cmd/merch_store

run: build
	./.bin/${BINARY_NAME}

clean: db-script-clean test-db-script-clean
	rm -rf .bin

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
