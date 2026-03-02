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


## Удаление созданных файлов
clean:
```bash
rm loglinter
rm loglinter.so
go clean -testcache
```