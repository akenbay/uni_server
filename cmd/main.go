package main

import (
	"context"
	"log"
	"os"
	"university/internal/handler"
	"university/internal/server"
	"university/internal/service"
	"university/internal/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// @title           University API
// @version         1.0
// @description     University management API - students, schedules, attendance
// @host            uni-server-29pn.onrender.com
// @schemes         https
// @BasePath        /
func main() {
	var err error

	encoderCfg := zap.NewProductionConfig()
	encoderCfg.EncoderConfig.TimeKey = "timestamp"
	encoderCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l, err := encoderCfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	logger := l.Sugar()
	defer logger.Sync()

	// Load .env file if it exists (for local development)
	// In production (Render.com), environment variables are set directly
	_ = godotenv.Load()

	db_url := os.Getenv("DATABASE_URL")
	if db_url == "" {
		logger.Fatal("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(context.Background(), db_url)
	if err != nil {
		logger.Fatal("Error connecting to database: ", err)
	}
	defer pool.Close()

	repo := storage.NewRepository(pool)

	repo.InitDB()

	svc := service.NewService(repo)
	hnd := handler.NewHandler(svc)
	srv := server.NewServer(hnd)

	// Use PORT environment variable (set by Render.com) or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof("Starting server on port %s", port)
	err = srv.Start(":" + port)
	if err != nil {
		logger.Fatal("server error: ", err)
	}
}
