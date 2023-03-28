package cli

import (
	"context"
	"io"

	"github.com/aereal/frontier"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
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
	app.base.Commands = append(app.base.Commands, &cliv2.Command{
		Name:   "deploy",
		Flags:  []cliv2.Flag{flagConfigPath, flagPublish},
		Action: app.actionDeploy,
	})
	return app
}

var (
	flagConfigPath = &cliv2.PathFlag{
		Name:  "config",
		Usage: "config file path",
		Value: cliv2.Path("function.yml"),
	}
	flagPublish = &cliv2.BoolFlag{
		Name:  "publish",
		Usage: "whether publish the function immediately",
		Value: true,
	}
)

func (a *App) actionDeploy(cliCtx *cliv2.Context) error {
	ctx := cliCtx.Context
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := cloudfront.NewFromConfig(cfg)
	configPath := cliCtx.Path(flagConfigPath.Name)
	doPublish := cliCtx.Bool(flagPublish.Name)
	deployer := frontier.NewDeployer(client, a.logger)
	return deployer.Deploy(ctx, configPath, doPublish)
}

func (a *App) Run(ctx context.Context, args []string) error {
	return a.base.RunContext(ctx, args)
}
