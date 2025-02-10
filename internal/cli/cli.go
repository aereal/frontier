//go:generate go run go.uber.org/mock/mockgen -build_constraint !live -typed -write_command_comment=false -write_package_comment=false -write_source_comment=false -package cli -destination ./mock_gen.go github.com/aereal/frontier/internal/cli DeployController,ImportController,RenderController

package cli

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/aereal/frontier"
	cli "github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type ImportController interface {
	Import(ctx context.Context, functionName string, configStream io.Writer, functionStream *frontier.WritableFile) error
}

type DeployController interface {
	Deploy(ctx context.Context, configPath string, publish bool) error
}

type RenderController interface {
	Render(ctx context.Context, configPath string, output io.Writer) error
}

type Controllers struct {
	ImportController
	DeployController
	RenderController
}

func New(input io.Reader, output, errOutput io.Writer, controllers Controllers) *App {
	return &App{
		input:         input,
		output:        output,
		errOutput:     errOutput,
		controllers:   controllers,
		shouldPublish: true,
	}
}

type App struct {
	input         io.Reader
	output        io.Writer
	errOutput     io.Writer
	controllers   Controllers
	shouldPublish bool
}

func (a *App) Run(ctx context.Context, args []string) error {
	rootCmdName := filepath.Base(args[0])
	cmd := &cli.Command{
		Name:      rootCmdName,
		Reader:    a.input,
		Writer:    a.output,
		ErrWriter: a.errOutput,
		Flags: []cli.Flag{
			flagOtelTraceEndpoint,
			flagLogLevel,
		},
		Before: a.onBefore,
		After:  a.onAfter,
		Commands: []*cli.Command{
			a.cmdRender(),
			a.cmdDeploy(),
			a.cmdImport(),
		},
	}
	for _, c := range cmd.Commands {
		instrumentTrace(c)
	}
	return cmd.Run(ctx, args)
}

func (a *App) onBefore(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if level := getLogLevel(cmd); level != slog.LevelInfo {
		swapLogger(func(prev *slog.Logger) *slog.Logger {
			lh := &leveledHandler{
				level:   level,
				Handler: prev.Handler(),
			}
			return slog.New(lh)
		})
	}
	if err := a.configureTracerProvider(ctx, cmd); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (a *App) onAfter(ctx context.Context, _ *cli.Command) error {
	if tp, ok := otel.GetTracerProvider().(interface{ Shutdown(context.Context) error }); ok {
		if err := tp.Shutdown(ctx); err != nil {
			slog.WarnContext(ctx, "failed to shutdown tracer provider", slog.String("error", err.Error()))
		}
	}
	return nil
}

func (a *App) configureTracerProvider(ctx context.Context, cmd *cli.Command) error { //nolint:unparam
	endpoint := cmd.String(flagOtelTraceEndpoint.Name)
	if endpoint == "" {
		return nil
	}

	slog.InfoContext(ctx, "set OTel trace endpoint", slog.String("endpoint", endpoint))

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint(endpoint))
	if err != nil {
		slog.WarnContext(ctx, "failed to setup otlptracegrpc", slog.String("error", err.Error()))
		return nil
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cmd.Name),
		semconv.ServiceVersion(cmd.Version),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return nil
}

var (
	flagConfigPath = &cli.StringFlag{
		Name:  "config",
		Usage: "config file path",
		Value: "function.yml",
	}
	flagOtelTraceEndpoint = &cli.StringFlag{
		Name:  "otel-trace-endpoint",
		Usage: "an endpoint (such as localhost:4317) to send OpenTelemetry traces. an empty value indicates no trace should be sent.",
	}
	flagLogLevel = &cli.GenericFlag{
		Name:  "log-level",
		Usage: "specify minimum log level. accepts valid [slog.Level] string representation.",
		Value: &logLevel{slog.LevelInfo},
	}
)
