package frontier

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

type Function struct {
	Name   string
	Code   *FunctionCode
	Config *FunctionConfig
}

type FunctionCode struct {
	Path string
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
	Comment string
	Runtime types.FunctionRuntime
}
