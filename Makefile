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

# Очистка
clean:
	rm -rf bin