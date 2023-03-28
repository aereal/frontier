package cli

import (
	"context"
	"io"

	cliv2 "github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

type App struct {
	base   *cliv2.App
	logger *zap.Logger
}

func New(logger *zap.Logger, input io.Reader, out, errOut io.Writer) *App {
	app := &App{
		logger: zap.NewNop(),
		base: &cliv2.App{
			Reader:    input,
			Writer:    out,
			ErrWriter: errOut,
		},
	}
	if logger != nil {
		app.logger = logger
	}
	return app
}

func (a *App) Run(ctx context.Context, args []string) error {
	return a.base.RunContext(ctx, args)
}
