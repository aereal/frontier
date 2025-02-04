package frontier

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func NewImporter(clientProvider CloudFrontClientProvider) *Importer {
	return &Importer{
		clientProvider: clientProvider,
	}
}

type Importer struct {
	clientProvider CloudFrontClientProvider
}

func (i *Importer) Import(ctx context.Context, functionName string, configStream io.Writer, functionStream *WritableFile) error {
	client, err := i.clientProvider.ProvideCloudFrontClient(ctx)
	if err != nil {
		return err
	}
	getInput := &cloudfront.GetFunctionInput{
		Name: &functionName,
	}
	getOut, err := client.GetFunction(ctx, getInput)
	if err != nil {
		noSuchFn := new(types.NoSuchFunctionExists)
		if errors.As(err, &noSuchFn) {
			return noSuchFn
		}
		return fmt.Errorf("GetFunction: %w", err)
	}

	describeInput := &cloudfront.DescribeFunctionInput{
		Name: &functionName,
	}
	describeOut, err := client.DescribeFunction(ctx, describeInput)
	if err != nil {
		return fmt.Errorf("DescribeFunction: %w", err)
	}

	if _, err := functionStream.Write(getOut.FunctionCode); err != nil {
		return err
	}
	fnCfg := &FunctionConfig{
		Comment: *describeOut.FunctionSummary.FunctionConfig.Comment,
		Runtime: describeOut.FunctionSummary.FunctionConfig.Runtime,
	}
	fn := &Function{
		Name:   *describeOut.FunctionSummary.Name,
		Config: fnCfg,
		Code: &FunctionCode{
			Path: functionStream.FilePath,
		},
	}
	if err := writeFunctionToStream(fn, configStream); err != nil {
		return err
	}
	return nil
}

type WritableFile struct {
	io.Writer
	FilePath string
}
