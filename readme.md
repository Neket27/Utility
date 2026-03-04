### Вы можете собрать линтер как исполняемый файл или как плагин для `golangci-lint`.

1.  **Клонируйте репозиторий:**
    ```bash
    git clone https://github.com/Neket27/Utility.git
    ```

2. **Скачайте зависимости:**
    ```bash
    go mod download
    ```
3.  **Соберите исполняемый файл:**
    ```bash
    go build -o loglinter ./cmd/loglinter
    ```
4. **Сделать файл исполняемым (Linux/Mac)**
    ```bash
   chmod +x ./loglinter
    ```

4.  **Соберите плагин для `golangci-lint`:**
    ```bash
    go build -o loglinter.so -buildmode=plugin ./cmd/loglinter
    ```


## Использование

### Как исполняемый файл

Вы можете запустить линтер напрямую для анализа пакета, standalone режим (для CI/CD).
Запуск линтера на тестовом файле:
```bash
./loglinter -config loglinter.yml ./test_app/main.go
```

## Пример работы
### Тестовый файл
```
package main

import (
	"go.uber.org/zap"
	"log/slog"
)

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	zapDevLogger, _ := zap.NewDevelopment()
	defer zapDevLogger.Sync()
	zapSugar := zapLogger.Sugar()
	defer zapSugar.Sync()

	slog.Info("starting server")           // обычное сообщение
	slog.Info("request to /api/v1/users")  // путь
	slog.Info("redirect to https://x.com") // URL
	slog.Info("")                          // пустая строка
	slog.Info("items 123 processed")       // числа допустимы
	slog.Info("apikey exposed")            // нет apikey
	slog.Info("secretive behavior")        // не whole word secret
	slog.Info("refresh token expired")

	slog.Info("user password: 123")
	slog.Info("secret: value")
	slog.Info("api_key=xyz")

	slog.Info("processing user_123")
	slog.Info("contact 123-456-7890")

	slog.Info("user_123 password reset")

	slog.Info("tokenized request") // не должно ловиться при whole word
	slog.Info("user_")             // нет цифр
	slog.Info("123-45-6789")       // должно начинаться с буквы
	slog.Info("   ")               // только пробелы
}

```
### Результат работы
[result_of_the_linter_work](result_of_the_linter_work)
```
~/❯ kulga-nikita@pc ./loglinter -config loglinter.yml ./test_app/main.go                                               kulga-nikita@pc
/home/neket/GolandProjects/utility/test_app/main.go:25:12: log message contains sensitive data: password
/home/neket/GolandProjects/utility/test_app/main.go:26:12: log message contains sensitive data: secret
/home/neket/GolandProjects/utility/test_app/main.go:27:12: log message contains sensitive data: api_key
/home/neket/GolandProjects/utility/test_app/main.go:29:12: log message matches forbidden pattern: user_\d+
/home/neket/GolandProjects/utility/test_app/main.go:30:12: log message matches forbidden pattern: \d{3}-\d{3}-\d{4}
/home/neket/GolandProjects/utility/test_app/main.go:32:12: log message matches forbidden pattern: user_\d+
/home/neket/GolandProjects/utility/test_app/main.go:36:12: log message must start with a lowercase letter
/home/neket/GolandProjects/utility/test_app/main.go:37:12: log message must start with a lowercase letter
```

### Линтер поддерживает гибкую настройку правил через YAML-файл.
По умолчанию (без передачи конфигурации) линтер будет использовать настройки по умолчанию.
```
rules:
  lowercase:
    enabled: true
    auto_fix_enabled: true

  english_only:
    enabled: true

  no_special_chars:
    enabled: true
    auto_fix_enabled: true
    max_consecutive_dots: 0

  sensitive_words:
    enabled: true
    auto_fix_enabled: true
    words:
      - password
      - passwd
      - pwd
      - secret
      - token
      - api_key
      - apikey
      - credential
      - private_key
      - access_token
      - refresh_token
      - secret_key
      - encryption_key
      - secret_token

    safe_phrases:
      - validated
      - verified
      - expired
      - refreshed
      - rotated
      - changed
      - updated
      - deleted
      - created
      - generated
      - revoked
      - invalid
      - missing
      - required
      - configured
      - initialized
      - loaded
      - saved
      - cleared
      - reset
      - processed
      - synchronized
      - exposed

  custom_patterns:
    enabled: true
    patterns:
      - 'user_\d+'
      - '\d{3}-\d{3}-\d{4}'

  # Правило 5: Кастомные regex-паттерны
  custom_patterns:
    enabled: true
    patterns:
      - 'user_\d+'              # Пример: user_12345
      - '\d{3}-\d{3}-\d{4}'     # Пример: телефон 123-456-7890
```

### Для исправления ошибок можно использовать команду:
```bash
./loglinter -config loglinter.yml --fix ./test_app/main.go
```
После чего тестовый файл будет изменён (приватные данные будут скрыты)
```azure
	slog.Info("starting server")          // обычное сообщение
	slog.Info("request to /api/v1/users") // путь
	slog.Info("redirect to https://xcom") // URL
	slog.Info("")                         // пустая строка
	slog.Info("items 123 processed")      // числа допустимы
	slog.Info("apikey exposed")           // нет apikey
	slog.Info("secretive behavior")       // не whole word secret
	slog.Info("refresh token expired")

	slog.Info("user password=[REDACTED]")
	slog.Info("secret=[REDACTED]")
	slog.Info("api_key=[REDACTED]")

	slog.Info("processing [REDACTED]")
	slog.Info("contact [REDACTED]")

	slog.Info("[REDACTED] password reset")

	slog.Info("tokenized request") // не должно ловиться при whole word
	slog.Info("user_")             // нет цифр
	slog.Info("123-45-6789")       // должно начинаться с буквы
	slog.Info("   ")               // только пробелы
```

## Удаление созданных файлов
clean:
```bash
rm loglinter
rm loglinter.so
go clean -testcache
```
