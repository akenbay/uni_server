package main

import (
	"log"
	"time"

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

}
