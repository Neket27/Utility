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

	// Edge cases (границы правил)
	slog.Info("hello world")             // ✅ Пробел допустим // TODO проверить
	slog.Debug("access token refreshed") // ✅  token //TODO проверить
	slog.Info("api_key: xyz789")         // ❌ api_key //TODO проверить
	slog.Debug("Token With Capital")     // ❌ token (case insensitive) //TODO проверить
	slog.Info("api request completed")   // ✅ api без key // TODO проверить

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
	slog.Info("items 123 processed")             // числа разрешены
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

	slog.Debug("access token refreshed") // Должно ловиться (нет safe phrase) || Не должно (есть safe phrase)
	slog.Info("token: abc123")           // Должно ловиться (опасный контекст)
	slog.Info("token validated")         // Должно ловиться (нет safe phrase) || Не должно (есть safe phrase)
	slog.Info("value: password123")      // Должно ловиться (опасный контекст)

}
