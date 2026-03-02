# Имя бинарника
BINARY=bin/loglinter
CMD_PATH=./cmd/loglinter
TEST_PATH=./test_app/...

.PHONY: all build chmod test run clean

# Основная цель
all: build chmod

# 1. Сборка бинарника линтера
build:
	go build -o $(BINARY) $(CMD_PATH)

# 2. Проверка прав доступа (Linux/Mac)
chmod:
	chmod +x $(BINARY)

# 3. Запуск линтера на тестовом файле
run: build chmod
	./$(BINARY) $(TEST_PATH)

# Очистка
clean:
	rm -rf bin