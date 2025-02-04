package cli

import (
	"context"

	"github.com/urfave/cli/v3"
)

func (a *App) cmdDeploy() *cli.Command {
	flagPublish := &cli.BoolFlag{
		Name:  "publish",
		Usage: "whether publish the function immediately",
		Value: a.shouldPublish,
		Action: func(_ context.Context, _ *cli.Command, val bool) error {
			a.shouldPublish = val
			return nil
		},
	}
	flagNoPublish := &cli.BoolFlag{
		Name:   "no-publish",
		Hidden: true,
		Value:  !a.shouldPublish,
		Action: func(_ context.Context, _ *cli.Command, val bool) error {
			a.shouldPublish = !val
			return nil
		},
	}
	return &cli.Command{
		Name: "deploy",
		// [cli.BoolWithInverseFlag] cannot have a default value because it causes `cannot set both flags` error.
		// refs. https://github.com/urfave/cli/blob/435b91c5099cfcf0fc4148b93e75ec9cc1a20b5c/flag_bool_with_inverse.go#L44C22-L44C28
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Flags: [][]cli.Flag{
					{
						flagPublish,
					},
					{
						flagNoPublish,
					},
				},
			},
		},
		Flags: []cli.Flag{
			flagConfigPath,
		},
		Action: a.actionDeploy,
	}
}

func (a *App) actionDeploy(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String(flagConfigPath.Name)
	doPublish := a.shouldPublish
	return a.controllers.Deploy(ctx, configPath, doPublish)
}
