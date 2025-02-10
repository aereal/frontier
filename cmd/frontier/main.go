package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/controller/listdist"
	"github.com/aereal/frontier/internal/cf"
	"github.com/aereal/frontier/internal/cli"
	"github.com/aereal/frontier/internal/fnarn"
)

func main() {
	os.Exit(run())
}

func run() int {
	sh := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})
	sl := slog.New(sh)
	slog.SetDefault(sl)
	var cfBuilder cf.SDKProvider
	arnResolver := fnarn.NewResolver(cfBuilder)
	controllers := cli.Controllers{
		RenderController:            frontier.NewRenderer(),
		ImportController:            frontier.NewImporter(cfBuilder),
		DeployController:            frontier.NewDeployer(cfBuilder),
		ListDistributionsController: listdist.NewController(cfBuilder),
	}
	if err := cli.New(os.Stdin, os.Stdout, os.Stderr, controllers, arnResolver).Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error(), slog.String("error", err.Error()))
		return 1
	}
	return 0
}
