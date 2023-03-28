package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/aereal/frontier"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cliv2 "github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type App struct {
	base   *cliv2.App
	logger *zap.Logger
}

const tracerName = "github.com/aereal/frontier/internal/cli"

func New(logger *zap.Logger, input io.Reader, out, errOut io.Writer) *App {
	app := &App{
		logger: zap.NewNop(),
		base: &cliv2.App{
			Reader:    input,
			Writer:    out,
			ErrWriter: errOut,
			Flags:     []cliv2.Flag{flagOtelTrace},
			Before: func(cliCtx *cliv2.Context) error {
				if !cliCtx.Bool(flagOtelTrace.Name) {
					return nil
				}

				ctx := cliCtx.Context
				exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
				if err != nil {
					logger.Warn("failed to setup otlptracegrpc", zap.Error(err))
					return nil
				}
				res := resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceVersion(cliCtx.App.Version),
					semconv.ServiceName(cliCtx.App.Name))
				tp := sdktrace.NewTracerProvider(
					sdktrace.WithBatcher(exporter),
					sdktrace.WithResource(res),
				)
				otel.SetTracerProvider(tp)
				return nil
			},
			After: func(cliCtx *cliv2.Context) error {
				if !cliCtx.Bool(flagOtelTrace.Name) {
					return nil
				}

				ctx := cliCtx.Context
				tp, ok := otel.GetTracerProvider().(interface{ Shutdown(context.Context) error })
				if ok {
					if err := tp.Shutdown(ctx); err != nil {
						logger.Warn("failed to shutdown tracer provider", zap.Error(err))
					}
				}
				return nil
			},
		},
	}
	if logger != nil {
		app.logger = logger
	}
	cmdDeploy := &cliv2.Command{
		Name:   "deploy",
		Flags:  []cliv2.Flag{flagConfigPath, flagPublish},
		Action: app.actionDeploy,
	}
	instrumentTrace(cmdDeploy)
	app.base.Commands = append(app.base.Commands, cmdDeploy)
	return app
}

var (
	flagOtelTrace = &cliv2.BoolFlag{
		Name:  "otel-trace",
		Usage: "enable OpenTelemetry traces",
		Value: false,
	}
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
	otelaws.AppendMiddlewares(&cfg.APIOptions)
	client := cloudfront.NewFromConfig(cfg)
	configPath := cliCtx.Path(flagConfigPath.Name)
	doPublish := cliCtx.Bool(flagPublish.Name)
	deployer := frontier.NewDeployer(client, a.logger)
	return deployer.Deploy(ctx, configPath, doPublish)
}

func (a *App) Run(ctx context.Context, args []string) error {
	return a.base.RunContext(ctx, args)
}

func instrumentTrace(cmd *cliv2.Command) {
	cmd.Before = func(cliCtx *cliv2.Context) error {
		ctx := cliCtx.Context
		cliCtx.Context, _ = otel.GetTracerProvider().Tracer(tracerName).Start(ctx, fmt.Sprintf("cli.%s", cliCtx.Command.FullName()))
		return nil
	}
	cmd.After = func(cliCtx *cliv2.Context) error {
		ctx := cliCtx.Context
		span := trace.SpanFromContext(ctx)
		span.End()
		return nil
	}
}
