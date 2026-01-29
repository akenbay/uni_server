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

	// Initialize database schema and seed data
	if err := initializeDatabase(context.Background(), db, logger); err != nil {
		logger.Fatal("Error initializing database: ", err)
	}

	repo := storage.NewRepository(db)
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

// initializeDatabase checks if database is initialized and runs init.sql if needed
func initializeDatabase(ctx context.Context, db *pgx.Conn, logger *zap.SugaredLogger) error {
	// Check if users table exists
	var tableExists bool
	err := db.QueryRow(ctx,
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')",
	).Scan(&tableExists)

	if err != nil {
		return err
	}

	if tableExists {
		logger.Info("Database already initialized")
		return nil
	}

	logger.Info("Initializing database from init.sql...")

	// Read init.sql file
	sqlBytes, err := os.ReadFile("init.sql")
	if err != nil {
		// Try parent directory
		sqlBytes, err = os.ReadFile("../init.sql")
		if err != nil {
			return err
		}
	}

	// Execute the entire SQL file
	_, err = db.Exec(ctx, string(sqlBytes))
	if err != nil {
		logger.Warnf("Database initialization warning: %v", err)
		// Continue anyway - some errors (like duplicate inserts) are acceptable
	}

	logger.Info("Database initialization completed")
	return nil
}
