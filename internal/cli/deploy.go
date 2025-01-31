package cli

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (a *App) cmdDeploy() *cli.Command {
	return &cli.Command{
		Name: "deploy",
		Flags: []cli.Flag{
			flagConfigPath,
			&cli.BoolWithInverseFlag{
				BoolFlag: &cli.BoolFlag{
					Name:  "publish",
					Usage: "whether publish the function immediately",
					Value: true,
				},
			},
		},
		Action: a.actionDeploy,
	}
}

func (a *App) actionDeploy(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String(flagConfigPath.Name)
	doPublish := cmd.Bool("publish")
	return a.controllers.Deploy(ctx, configPath, doPublish)
}
