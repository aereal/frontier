package cli

import (
	"context"
	"fmt"

	cli "github.com/urfave/cli/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "github.com/aereal/frontier/internal/cli"
)

func instrumentTrace(cmd *cli.Command) {
	cmd.Before = prependBeforeFunc(cmd.Before, func(ctx context.Context, c *cli.Command) (context.Context, error) {
		spanName := fmt.Sprintf("cli.%s", cmd.FullName())
		ctx, _ = otel.GetTracerProvider().Tracer(tracerName).Start(ctx, spanName)
		return ctx, nil
	})
	cmd.After = appendAfterFunc(cmd.After, func(ctx context.Context, _ *cli.Command) error {
		trace.SpanFromContext(ctx).End()
		return nil
	})
}

func prependBeforeFunc(current, prepended cli.BeforeFunc) cli.BeforeFunc {
	if current == nil {
		return prepended
	}
	return func(ctx context.Context, c *cli.Command) (context.Context, error) {
		ctx, err := prepended(ctx, c)
		if err != nil {
			return ctx, err
		}
		return current(ctx, c)
	}
}

func appendAfterFunc(prev, appended cli.AfterFunc) cli.AfterFunc {
	if prev == nil {
		return appended
	}
	return func(ctx context.Context, c *cli.Command) error {
		if err := prev(ctx, c); err != nil {
			return err
		}
		return appended(ctx, c)
	}
}
