package example

import (
	"log/slog"
)

func testLogMessages() {

	slog.Info("starting server on port 8080")
	slog.Debug("database connection established")
	slog.Warn("cache miss for key user_123")
	slog.Error("failed to connect to database")

	slog.Info("Starting server")             // want "log message must start with a lowercase letter"
	slog.Debug("Database connection failed") // want "log message must start with a lowercase letter"

	slog.Info("запуск сервера")      // want "log message must be in English only"
	slog.Error("ошибка подключения") // want "log message must be in English only"

	slog.Info("server started!")         // want "log message must not contain special characters or emojis"
	slog.Error("connection failed!!")    // want "log message must not contain special characters or emojis"
	slog.Warn("something went wrong...") // want "log message must not contain special characters or emojis"

	password := "secret123"
	apiKey := "key123"
	slog.Info("user password: " + password) // want "log message contains sensitive data: password"
	slog.Debug("api_key=" + apiKey)         // want "log message contains sensitive data: api_key"

	slog.Info("Starting server") // want "log message must start with a lowercase letter"
	slog.Info("server started!") // want "log message must not contain special characters or emojis"
}
