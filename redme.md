Вы можете собрать линтер как исполняемый файл или как плагин для `golangci-lint`.

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

### Линтер поддерживает гибкую настройку правил через YAML-файл. 
По умолчанию (без передачи конфигурации) линтер будет использовать настройки по умолчанию.
```bash
rules:
  # Правило 1: Сообщение должно начинаться со строчной буквы
  lowercase:
    enabled: true  # false — отключить правило

  # Правило 2: Только английский язык (Latin script)
  english_only:
    enabled: true

  # Правило 3: Запрет спецсимволов и эмодзи
  no_special_chars:
    enabled: true

  # Правило 4: Проверка на чувствительные данные
  sensitive_words:
    enabled: true
    # Список слов для поиска (регистронезависимый)
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
    # Фразы, которые делают использование слова безопасным
    # Например: "password validated" — OK, "password: 123" — ошибка
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

  # Правило 5: Кастомные regex-паттерны
  custom_patterns:
    enabled: true
    patterns:
      - 'user_\d+'              # Пример: user_12345
      - '\d{3}-\d{3}-\d{4}'     # Пример: телефон 123-456-7890
```

## Удаление созданных файлов
clean:
```bash
rm loglinter
rm loglinter.so
go clean -testcache
```