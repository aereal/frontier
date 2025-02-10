package frontier

import (
	"context"
	"io"
)

func NewRenderer() *Renderer {
	return &Renderer{}
}

type Renderer struct{}

func (r *Renderer) Render(ctx context.Context, configPath string, output io.Writer) error {
	fn, err := ParseConfigFromPath(configPath)
	if err != nil {
		return err
	}
	if err := writeFunctionToStream(fn, output); err != nil {
		return err
	}
	return nil
}
