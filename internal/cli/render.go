package cli

import (
	"context"

	"github.com/aereal/frontier"
	cli "github.com/urfave/cli/v3"
)

func (a *App) cmdRender() *cli.Command {
	return &cli.Command{
		Name:        "render",
		Description: "render resolved function config",
		Flags: []cli.Flag{
			flagConfigPath,
		},
		Action: a.actionRender,
	}
}

func (a *App) actionRender(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String(flagConfigPath.Name)
	renderer := frontier.NewRenderer()
	return renderer.Render(ctx, configPath, cmd.Writer)
}
