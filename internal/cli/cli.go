package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"

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
)

type App struct {
	base *cliv2.App
}

const tracerName = "github.com/aereal/frontier/internal/cli"

func New(input io.Reader, out, errOut io.Writer) *App {
	app := &App{
		base: &cliv2.App{
			Reader:    input,
			Writer:    out,
			ErrWriter: errOut,
			Flags:     []cliv2.Flag{flagOtelTrace, flagOtelTraceEndpoint, flagLogLevel},
			Before: func(cliCtx *cliv2.Context) error {
				if level := getLogLevel(cliCtx); level != slog.LevelInfo {
					lh := &leveledHandler{
						level: level, Handler: slog.Default().Handler(),
					}
					slog.SetDefault(slog.New(lh))
				}

				endpoint := cliCtx.String(flagOtelTraceEndpoint.Name)
				ctx := cliCtx.Context
				if endpoint == "" {
					if !cliCtx.Bool(flagOtelTrace.Name) {
						return nil
					}
					endpoint = defaultOtelTraceEndpoint
				}

				exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint(endpoint))
				if err != nil {
					slog.WarnContext(ctx, "failed to setup otlptracegrpc", slog.String("error", err.Error()))
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
				if !cliCtx.Bool(flagOtelTrace.Name) || cliCtx.String(flagOtelTraceEndpoint.Name) == "" {
					return nil
				}

				ctx := cliCtx.Context
				tp, ok := otel.GetTracerProvider().(interface{ Shutdown(context.Context) error })
				if ok {
					if err := tp.Shutdown(ctx); err != nil {
						slog.WarnContext(ctx, "failed to shutdown tracer provider", slog.String("error", err.Error()))
					}
				}
				return nil
			},
		},
	}
	cmdDeploy := &cliv2.Command{
		Name:   "deploy",
		Flags:  []cliv2.Flag{flagConfigPath, flagPublish},
		Action: app.actionDeploy,
	}
	cmdRender := &cliv2.Command{
		Name:        "render",
		Flags:       []cliv2.Flag{flagConfigPath},
		Description: "render resolved function config",
		Action:      app.actionRender,
	}
	app.base.Commands = append(app.base.Commands, cmdDeploy, cmdRender)
	for _, c := range app.base.Commands {
		instrumentTrace(c)
	}
	return app
}

var (
	defaultOtelTraceEndpoint = "localhost:4317"

	flagOtelTraceEndpoint = &cliv2.StringFlag{
		Name:  "otel-trace-endpoint",
		Usage: "an endpoint (such as localhost:4317) to send OpenTelemetry traces. an empty value indicates no trace should be sent.",
	}
	flagOtelTrace = &cliv2.BoolFlag{
		Name:  "otel-trace",
		Usage: "Deprecated: use --otel-trace-endpoint. enable OpenTelemetry traces.",
		Value: false,
		Action: func(ctx *cliv2.Context, b bool) error {
			slog.WarnContext(ctx.Context, "--otel-trace option is deprecated. use --otel-trace-endpoint")
			return nil
		},
	}
	flagLogLevel = &cliv2.GenericFlag{
		Name:  "log-level",
		Usage: "specify minimum log level. accepts valid [slog.Level] string representation.",
		Value: &logLevel{slog.LevelInfo},
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
	deployer := frontier.NewDeployer(client)
	return deployer.Deploy(ctx, configPath, doPublish)
}

func (a *App) actionRender(cliCtx *cliv2.Context) error {
	configPath := cliCtx.Path(flagConfigPath.Name)
	renderer := frontier.NewRenderer(configPath, cliCtx.App.Writer)
	return renderer.Render(cliCtx.Context)
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

type logLevel struct {
	slog.Level
}

var _ cliv2.Generic = (*logLevel)(nil)

func (l *logLevel) Set(v string) error {
	if err := l.Level.UnmarshalText([]byte(v)); err != nil {
		return err
	}
	return nil
}

func getLogLevel(cliCtx *cliv2.Context) slog.Level {
	ll, ok := cliCtx.Generic(flagLogLevel.Name).(*logLevel)
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
