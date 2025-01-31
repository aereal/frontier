package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/internal/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	sh := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})
	sl := slog.New(sh)
	slog.SetDefault(sl)
	sdkCFProvider := cli.SDKCloudFrontClientProvider{}
	deployer := frontier.NewDeployer(sdkCFProvider)
	importer := frontier.NewImporter(sdkCFProvider)
	renderer := frontier.NewRenderer()
	if err := cli.New(os.Stdin, os.Stdout, os.Stderr, cli.GRPCExporterFactory{}, deployer, importer, renderer).Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error(), slog.String("error", err.Error()))
		return 1
	}
	return 0
}
