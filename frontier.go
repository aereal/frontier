package frontier

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

type Function struct {
	Name   string
	Code   string
	Config *FunctionConfig
}

func (f *Function) toCreateInput() *cloudfront.CreateFunctionInput {
	return &cloudfront.CreateFunctionInput{
		Name:         &f.Name,
		FunctionCode: []byte(f.Code),
		FunctionConfig: &types.FunctionConfig{
			Comment: &f.Config.Comment,
			Runtime: f.Config.Runtime,
		},
	}
}

func (fn *Function) toUpdateInput(etag *string) *cloudfront.UpdateFunctionInput {
	return &cloudfront.UpdateFunctionInput{
		Name:         &fn.Name,
		FunctionCode: []byte(fn.Code),
		IfMatch:      etag,
		FunctionConfig: &types.FunctionConfig{
			Comment: &fn.Config.Comment,
			Runtime: fn.Config.Runtime,
		},
	}
}

type FunctionConfig struct {
	Comment string
	Runtime types.FunctionRuntime
}
