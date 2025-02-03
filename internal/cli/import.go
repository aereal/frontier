package cli

import (
	"context"
	"os"

	"github.com/aereal/frontier"
	"github.com/urfave/cli/v3"
)

func (a *App) cmdImport() *cli.Command {
	return &cli.Command{
		Name: "import",
		Flags: []cli.Flag{
			flagConfigPath,
			&cli.StringFlag{
				Name:     "name",
				Usage:    "function name",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "function-path",
				Usage: "function implementation path",
				Value: "fn.js",
			},
		},
		Action: a.actionImport,
	}
}

func (a *App) actionImport(ctx context.Context, cmd *cli.Command) error {
	functionName := cmd.String("name")
	if functionName == "" {
		return ErrFunctionNameRequired
	}
	functionPath := cmd.String("function-path")
	if functionPath == "" {
		return ErrFunctionPathRequired
	}
	configPath := cmd.String(flagConfigPath.Name)
	if configPath == "" {
		return ErrConfigPathRequired
	}

	fnFile, err := openForWrite(functionPath, 0600)
	if err != nil {
		return err
	}
	defer fnFile.Close()
	configFile, err := openForWrite(configPath, 0600)
	if err != nil {
		return err
	}
	defer configFile.Close()

	functionOut := &frontier.WritableFile{
		FilePath: functionPath,
		Writer:   fnFile,
	}
	return a.controllers.Import(ctx, functionName, configFile, functionOut)
}

func openForWrite(name string, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}
