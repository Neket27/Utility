package main

import (
	"go.uber.org/zap"
	"log"
	"log/slog"
)

func main() {
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()

	zapDevLogger, _ := zap.NewDevelopment()
	defer zapDevLogger.Sync()

	zapSugar := zapLogger.Sugar()
	defer zapSugar.Sync()

	// ========================================================================
	// ✅ ПРАВИЛЬНЫЕ ВАРИАНТЫ (не должны ловиться линтером)
	// ========================================================================

	// Стандартный log
	log.Println("starting server on port 8080")
	log.Printf("server initialized")

	// Slog
	slog.Info("starting server")
	slog.Debug("debug information")
	slog.Warn("warning message")
	slog.Error("error occurred")

	// Zap
	zapLogger.Info("database connection established")
	zapLogger.Debug("query executed")
	zapLogger.Error("connection failed")

	// Zap Sugar
	zapSugar.Info("user authenticated successfully")
	zapSugar.Warn("cache miss")
	zapSugar.Error("request timeout")

	// Zap Dev Logger
	zapDevLogger.Info("development mode active")

	// ========================================================================
	// ❌ ПРАВИЛО 1: Заглавная буква в начале сообщения
	// ========================================================================

	// Slog violations
	slog.Info("Starting server")    // ❌ 'S' заглавная
	slog.Error("Failed to connect") // ❌ 'F' заглавная
	slog.Warn("Warning detected")   // ❌ 'W' заглавная

	// Log violations
	log.Println("Server started")        // ❌ 'S' заглавная
	log.Printf("Connection established") // ❌ 'C' заглавная

	// Zap violations
	zapLogger.Info("Database connection established") // ❌ 'D' заглавная
	zapLogger.Error("Query failed")                   // ❌ 'Q' заглавная
	zapSugar.Info("User logged in")                   // ❌ 'U' заглавная
	zapDevLogger.Warn("Configuration missing")        // ❌ 'C' заглавная

	// ========================================================================
	// ❌ ПРАВИЛО 2: Не английский язык (кириллица, иероглифы, etc.)
	// ========================================================================

	// Slog violations
	slog.Info("запуск сервера")                    // ❌ Кириллица
	slog.Error("ошибка подключения к базе данных") // ❌ Кириллица
	slog.Warn("предупреждение о памяти")           // ❌ Кириллица

	// Log violations
	log.Println("сервер запущен")    // ❌ Кириллица
	log.Printf("ошибка авторизации") // ❌ Кириллица

	// Zap violations
	zapLogger.Info("пользователь авторизован") // ❌ Кириллица
	zapSugar.Error("критическая ошибка")       // ❌ Кириллица

	// Mixed language (тоже должно ловиться)
	slog.Info("server запущен successfully") // ❌ Mixed

	// ========================================================================
	// ❌ ПРАВИЛО 3: Спецсимволы и эмодзи
	// ========================================================================

	// Emojis
	slog.Info("server started! 🚀")       // ❌ Эмодзи
	slog.Error("connection failed!!! 💥") // ❌ Эмодзи + спецсимволы
	zapLogger.Info("success! 🎉")         // ❌ Эмодзи

	// Multiple special characters
	slog.Info("server started!!!") // ❌ Множественные !
	slog.Warn("warning...")        // ❌ Троеточие
	slog.Error("error???")         // ❌ Множественные ?
	log.Println("done---")         // ❌ Множественные -

	// Other special chars
	//TODO посмотреть корректность
	zapSugar.Info("status: @ok")        // ❌ @ символ
	zapSugar.Debug("value: #123")       // ❌ # символ
	zapDevLogger.Info("path: /usr/bin") // ⚠️ / допустим в путях

	// Edge cases (границы правил)
	slog.Info("hello world")   // ✅ Пробел допустим // TODO проверить
	slog.Info("hello, world!") // ✅ , и ! допустимы по одному
	slog.Info("test-case")     // ✅ - допустим
	slog.Info("it's working")  // ✅ ' допустим

	// ========================================================================
	// ❌ ПРАВИЛО 4: Чувствительные данные (ключевые слова)
	// ========================================================================

	// Password variations
	slog.Info("user password: 123")        // ❌ password
	slog.Debug("password reset requested") // ✅  password
	log.Println("admin password changed")  // ✅  password

	// Token variations
	slog.Info("token: abc123")              // ❌ token
	slog.Debug("access token refreshed")    // ✅  token //TODO проверить
	zapLogger.Info("refresh token expired") // ❌ token

	// API Key variations
	slog.Debug("api_key=secret")    // ❌ api_key
	slog.Info("api_key: xyz789")    // ❌ api_key //TODO проверить
	zapSugar.Warn("apikey exposed") // ✅  apikey

	// Secret variations
	slog.Info("secret: confidential")    // ❌ secret
	zapLogger.Error("secret key leaked") // ❌ secret + key

	// Key variations
	slog.Debug("encryption key: aes256") // ❌ key
	log.Printf("private key loaded")     // ❌ key

	// Pass variations (сокращение от password)
	slog.Info("pass: temp123")     // ❌ pass
	zapSugar.Debug("pass changed") // ❌ pass

	// Edge cases (должны ловиться)
	slog.Info("PASSWORD in uppercase") // ❌ password (case insensitive)
	slog.Debug("Token With Capital")   // ❌ token (case insensitive) //TODO проверить

	slog.Info("user authenticated successfully") // ✅ Нет ключевых слов
	slog.Info("api request completed")           // ✅ api без key // TODO проверить
	slog.Info("token validated")                 // ✅  Граничный случай
	log.Println("secret sauce recipe")           //  ✅  Граничный случай (не про данные)

	// ========================================================================
	// ❌ КОМБИНИРОВАННЫЕ НАРУШЕНИЯ (несколько правил сразу)
	// ========================================================================

	// Заглавная + Кириллица
	slog.Info("Запуск сервера") // ❌ Правила 1 и 2

	// Заглавная + Спецсимволы
	slog.Error("FAILED!!!") // ❌ Правила 1 и 3

	// Заглавная + Чувствительные данные
	zapLogger.Info("Password: secret123") // ❌ Правила 1 и 4

	// Все 4 нарушения сразу
	slog.Error("PASSWORD: секрет123!!! 🔥") // ❌ Все 4 правила

	// Кириллица + Эмодзи
	slog.Warn("ошибка! 🚫") // ❌ Правила 2 и 3

	// Чувствительные + Спецсимволы
	zapSugar.Debug("api_key=secret!!!") // ❌ Правила 3 и 4

	// ========================================================================
	// ГРАНИЧНЫЕ СЛУЧАИ (edge cases)
	// ========================================================================

	// Пустые строки
	slog.Info("") // ✅  Пустое сообщение

	// Только пробелы
	slog.Info("   ") // ❌ Только пробелы

	// Очень длинное сообщение
	slog.Info("this is a very long log message that should still be processed correctly by the linter without any issues or performance degradation")

	// Unicode границы
	slog.Info("café")   // ✅  Latin-1 (может быть допустимо)
	slog.Info("Москва") // ❌ Кириллица

	// Числа в начале (должны быть допустимы)
	slog.Info("123 items processed") // ❌ Начинается с числа

	// Аббревиатуры
	slog.Info("HTTP request completed") // ❌ HTTP заглавные
	slog.Info("API response received")  // ❌ API заглавные

	// Пути и URL
	slog.Info("request to /api/v1/users")        // ✅ Пути допустимы
	slog.Info("redirect to https://example.com") // ✅ URL допустимы

	/////////////////////

	// ========================================================================
	// ✅ ДОЛЖНЫ ПРОХОДИТЬ (нет sensitive слов и regex)
	// ========================================================================

	slog.Info("Starting server")                 // uppercase разрешен
	slog.Info("сервер запущен")                  // кириллица разрешена
	slog.Info("error!!! 🚀")                      // спецсимволы разрешены
	slog.Info("123 items processed")             // числа разрешены
	slog.Info("user authenticated successfully") // ок
	slog.Info("api request completed")           // api без key — ок
	slog.Info("secret sauce recipe")             // нет точного совпадения secret? → зависит от реализации
	slog.Info("token validated")                 // если просто contains → сработает!

	// ========================================================================
	// ❌ sensitive_words (должны ловиться)
	// ========================================================================

	slog.Info("user password: 123")       // ❌ password
	slog.Info("PASSWORD reset requested") // ❌ password (case-insensitive)
	slog.Info("token: abc123")            // ❌ token
	slog.Info("refresh token expired")    // ❌ token
	slog.Info("api_key=xyz")              // ❌ api_key
	slog.Info("secret: value")            // ❌ secret
	slog.Info("my_custom_secret exposed") // ❌ my_custom_secret
	zapLogger.Info("secret key leaked")   // ❌ secret
	zapSugar.Debug("password changed")    // ❌ password

	// ========================================================================
	// ❌ custom_patterns (regex)
	// ========================================================================

	slog.Info("processing user_12345")   // ❌ user_\d+
	slog.Info("user_1 created")          // ❌ user_\d+
	slog.Info("contact 123-456-7890")    // ❌ phone pattern
	log.Println("call me 555-123-4567")  // ❌ phone pattern
	zapLogger.Info("user_999 logged in") // ❌ user_\d+

	// ========================================================================
	// ❌ Комбинированные (оба правила)
	// ========================================================================

	slog.Info("user_123 password reset")     // ❌ regex + sensitive
	slog.Info("user_777 token generated")    // ❌ regex + sensitive
	slog.Info("123-456-7890 secret exposed") // ❌ phone + sensitive

	// ========================================================================
	// 🧪 ГРАНИЧНЫЕ СЛУЧАИ (проверят реализацию)
	// ========================================================================

	slog.Info("apikey exposed")     // ❓ НЕ должно ловиться (нет api_key)
	slog.Info("secretive behavior") // ❓ зависит от contains или whole word
	slog.Info("us er_123")          // ❓ НЕ должно ловиться
	slog.Info("123-45-6789")        // ❓ НЕ совпадает с phone regex
	slog.Info("user_")              // ❓ нет цифр
	slog.Info("tokenized request")  // ❓ зависит от логики contains

}
