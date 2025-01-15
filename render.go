package frontier

import (
	"context"
	"io"

	"gopkg.in/yaml.v3"
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
	enc := yaml.NewEncoder(r.output)
	enc.SetIndent(2)
	if err := enc.Encode(fn); err != nil {
		return err
	}
	return nil
}
