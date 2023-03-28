package cli

import (
	"context"
	"io"

	cliv2 "github.com/urfave/cli/v2"
)

type App struct {
	base *cliv2.App
}

func New(input io.Reader, out, errOut io.Writer) *App {
	app := &App{
		base: &cliv2.App{
			Reader:    input,
			Writer:    out,
			ErrWriter: errOut,
		},
	}
	return app
}

func (a *App) Run(ctx context.Context, args []string) error {
	return a.base.RunContext(ctx, args)
}
