package cli

import (
	"context"

	"github.com/aereal/frontier"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/urfave/cli/v3"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
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
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	otelaws.AppendMiddlewares(&cfg.APIOptions)
	client := cloudfront.NewFromConfig(cfg)
	configPath := cmd.String(flagConfigPath.Name)
	doPublish := cmd.Bool("publish")
	deployer := frontier.NewDeployer(client)
	return deployer.Deploy(ctx, configPath, doPublish)
}
