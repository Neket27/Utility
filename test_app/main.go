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
