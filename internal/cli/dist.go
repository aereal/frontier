package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"slices"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/controller/listdist"
	"github.com/aereal/frontier/internal/fnarn"
	"github.com/aereal/frontier/internal/presenter"
	"github.com/aereal/frontier/internal/presenter/json"
	"github.com/urfave/cli/v3"
)

func (a *App) cmdDist() *cli.Command {
	return &cli.Command{
		Name:  "dist",
		Usage: "manage distribution",
		Commands: []*cli.Command{
			a.cmdDistList(),
		},
		Writer:    a.output,
		ErrWriter: a.errOutput,
		Reader:    a.input,
	}
}

func (a *App) cmdDistList() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "list distributions",
		Action:    a.actionDistList,
		Writer:    a.output,
		ErrWriter: a.errOutput,
		Reader:    a.input,
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Category: "search criteria",
				Flags: [][]cli.Flag{
					{
						&cli.StringFlag{
							Name:  "function-arn",
							Usage: "search distributions associated with functions identified by the value",
						},
					},
					{
						&cli.StringFlag{
							Name:  "function-name",
							Usage: "search distributions associated with functions named with the value",
						},
					},
					{
						&cli.BoolFlag{
							Name:  "current",
							Usage: "search distributions associated with the function defined in the config file",
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			flagConfigPath,
			&cli.FlagBase[OutputFormat, cli.NoConfig, outputFormatCreator]{
				Name:  "format",
				Usage: usageText(slices.Values(AvailableOutputFormatValues()), "output format"),
				Value: OutputFormatJSON,
			},
			&cli.StringFlag{
				Name:     "event-type",
				Usage:    "list only associations that run functions against given event type",
				Category: "search criteria",
			},
		},
	}
}

func (a *App) actionDistList(ctx context.Context, cmd *cli.Command) error {
	format, ok := cmd.Value("format").(OutputFormat)
	if !ok {
		format = OutputFormatJSON
	}

	var presenter presenter.AssociatedDistributionsPresenter
	switch format {
	case OutputFormatJSON:
		presenter = json.NewAssociatedDistributionsPresenter(cmd.Writer)
	case OutputFormatJSONPretty:
		presenter = json.NewAssociatedDistributionsPresenter(cmd.Writer, json.Pretty(true))
	}

	criteria := listdist.NewCriteria()
	if eventType := cmd.String("event-type"); eventType != "" {
		criteria.Add(listdist.EqualEventType(eventType))
	}
	if functionArn := cmd.String("function-arn"); functionArn != "" {
		criteria.Add(listdist.EqualFunctionArn(functionArn))
	}
	if functionName := cmd.String("function-name"); functionName != "" {
		functionArn, err := a.arnResolver.ResolveFunctionARN(ctx, fnarn.FunctionName(functionName))
		if err != nil {
			return err
		}
		criteria.Add(listdist.EqualFunctionArn(functionArn))
	}
	if cmd.Bool("current") {
		cfg, err := frontier.ParseConfigFromPath(cmd.String(flagConfigPath.Name))
		if err != nil {
			return err
		}
		functionArn, err := a.arnResolver.ResolveFunctionARN(ctx, fnarn.FunctionName(cfg.Name))
		if err != nil {
			return err
		}
		criteria.Add(listdist.EqualFunctionArn(functionArn))
	}
	associations, err := a.controllers.ListDistributions(ctx, cmd.Writer, criteria)
	if err != nil {
		return err
	}
	presenter.PresentAssociatedDistributions(associations)
	return nil
}

func join[T fmt.Stringer](out io.Writer, xs iter.Seq[T], sep string) {
	var seen bool
	next, stop := iter.Pull(xs)
	defer stop()
	for {
		str, ok := next()
		if !ok {
			break
		}
		if seen {
			fmt.Fprint(out, sep)
		}
		fmt.Fprint(out, str.String())
		seen = true
	}
}

func usageText[T fmt.Stringer](choices iter.Seq[T], usage string) string {
	buf := new(bytes.Buffer)
	join(buf, choices, ", ")
	return fmt.Sprintf("%s (available values: %s)", usage, buf)
}
