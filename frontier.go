package frontier

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"gopkg.in/yaml.v3"
)

func parseConfigFromPath(configPath string) (*Function, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()
	fn := new(Function)
	if err := yaml.NewDecoder(f).Decode(fn); err != nil {
		return nil, fmt.Errorf("yaml.Decoder.Decode: %w", err)
	}
	return fn, nil
}

type Function struct {
	Name   string          `yaml:"name"`
	Code   *FunctionCode   `yaml:"code"`
	Config *FunctionConfig `yaml:"config"`
}

type FunctionCode struct {
	Path string `yaml:"path"`
}

func (f *Function) toCreateInput() (*cloudfront.CreateFunctionInput, error) {
	body, err := os.ReadFile(f.Code.Path)
	if err != nil {
		return nil, err
	}
	return &cloudfront.CreateFunctionInput{
		Name:         &f.Name,
		FunctionCode: body,
		FunctionConfig: &types.FunctionConfig{
			Comment: &f.Config.Comment,
			Runtime: f.Config.Runtime,
		},
	}, nil
}

func (fn *Function) toUpdateInput(etag *string) (*cloudfront.UpdateFunctionInput, error) {
	body, err := os.ReadFile(fn.Code.Path)
	if err != nil {
		return nil, err
	}
	return &cloudfront.UpdateFunctionInput{
		Name:         &fn.Name,
		FunctionCode: body,
		IfMatch:      etag,
		FunctionConfig: &types.FunctionConfig{
			Comment: &fn.Config.Comment,
			Runtime: fn.Config.Runtime,
		},
	}, nil
}

type FunctionConfig struct {
	Comment string                `yaml:"comment"`
	Runtime types.FunctionRuntime `yaml:"runtime"`
}
