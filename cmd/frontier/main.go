package main

import (
	"context"
	"os"

	"github.com/aereal/frontier/internal/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger, err := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		Encoding:         "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "severity",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
		},
	}.Build()
	if err != nil {
		return 1
	}
	defer func() {
		_ = logger.Sync()
	}()
	if err := cli.New(logger, os.Stdin, os.Stdout, os.Stderr).Run(context.Background(), os.Args); err != nil {
		logger.Error(err.Error(), zap.Error(err))
		return 1
	}
	return 0
}
