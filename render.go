package frontier

import (
	"context"
	"io"
)

func NewRenderer(configPath string, output io.Writer) *Renderer {
	return &Renderer{
		configPath: configPath,
		output:     output,
	}
}

type Renderer struct {
	configPath string
	output     io.Writer
}

func (r *Renderer) Render(ctx context.Context) error {
	fn, err := parseConfigFromPath(r.configPath)
	if err != nil {
		return err
	}
	if err := writeFunctionToStream(fn, r.output); err != nil {
		return err
	}
	return nil
}
