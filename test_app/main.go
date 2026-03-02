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

	slog.Debug("access token refreshed") // ❌ Должно ловиться (нет safe phrase) ✅ Не должно (есть safe phrase)
	slog.Info("token: abc123")           // ❌ Должно ловиться (опасный контекст)
	slog.Info("token validated")         // ❌ Должно ловиться (нет safe phrase) ✅ Не должно (есть safe phrase)

}
