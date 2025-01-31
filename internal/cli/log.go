package cli

import (
	"context"
	"log/slog"
	"sync"

	"github.com/urfave/cli/v3"
)

type logLevel struct {
	slog.Level
}

var _ cli.Value = (*logLevel)(nil)

func (l logLevel) String() string { return l.Level.String() }

func (l *logLevel) Set(v string) error {
	if err := l.Level.UnmarshalText([]byte(v)); err != nil {
		return err
	}
	return nil
}

func (l logLevel) Get() any { return l.Level }

func getLogLevel(cmd *cli.Command) slog.Level {
	ll, ok := cmd.Value(flagLogLevel.Name).(*logLevel)
	if ok {
		return ll.Level
	}
	return slog.LevelInfo
}

type leveledHandler struct {
	slog.Handler
	level slog.Level
}

var _ slog.Handler = (*leveledHandler)(nil)

func (h *leveledHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

var loggerMutex sync.Mutex

func swapLogger(swapFn func(prev *slog.Logger) *slog.Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	newLogger := swapFn(slog.Default())
	slog.SetDefault(newLogger)
}
