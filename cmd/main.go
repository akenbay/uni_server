package main

import (
	"context"
	"log"
	"os"
	"time"
	"university/internal/handler"
	"university/internal/server"
	"university/internal/service"
	"university/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var err error
	time.Local, err = time.LoadLocation("Asia/Almaty")
	if err != nil {
		log.Fatal(err)
	}

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

	db, err := pgx.Connect(context.Background(), db_url)
	if err != nil {
		logger.Fatal("Error connecting to database: ", err)
	}
	defer db.Close(context.Background())

	repo := storage.NewRepository(db)

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
