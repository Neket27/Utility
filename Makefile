BINARY=bin/loglinter
CMD_PATH=./cmd/loglinter
TEST_PATH=./test_app/...
CONFIG_PATH=./loglinter.yml

.PHONY: all build chmod test run clean

all: build chmod

# 1. Сборка бинарника линтера
build:
	go build -o $(BINARY) $(CMD_PATH)

# 2. Сделать файл исполняемым (Linux/Mac)
chmod:
	chmod +x $(BINARY)

# 3. Запуск линтера на тестовом файле
run: build chmod
	./$(BINARY) -config $(CONFIG_PATH) $(TEST_PATH)

# 4. Сборка плагина
build-plugin:
	go build -o loglinter.so -buildmode=plugin $(CMD_PATH)

# Очистка
clean:
	rm -rf bin
	rm loglinter.so
	go clean -testcache




# Дополнительные цели

# Установка golangci-lint для тестов
install-golangci:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Тесты
test:
	go test -cover ./pkg/...


