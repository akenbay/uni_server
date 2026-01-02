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

	err = godotenv.Load()

	if err != nil {
		logger.Fatal("Error loading .env")
	}

	db_url := os.Getenv("DATABASE_URL")
	db, err := pgx.Connect(context.Background(), db_url)
	if err != nil {
		logger.Fatal("Error connecting to database: ", err)
	}
	defer db.Close(context.Background())

	repo := storage.NewRepository(db)
	svc := service.NewService(repo)
	hnd := handler.NewHandler(srv)
	srv := server.NewServer(hnd)

	err = server.Start(":8080")
	if err != nil {
		logger.Fatal("server error: ", err)
	}
}
