package frontier

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func NewImporter(client CFForImport, functionName string, configStream io.Writer, functionStream *WritableFile) *Importer {
	return &Importer{
		client:         client,
		functionName:   functionName,
		configStream:   configStream,
		functionStream: functionStream,
	}
}

type Importer struct {
	functionName   string
	client         CFForImport
	configStream   io.Writer
	functionStream *WritableFile
}

func (i *Importer) Import(ctx context.Context) error {
	getInput := &cloudfront.GetFunctionInput{
		Name: &i.functionName,
	}
	getOut, err := i.client.GetFunction(ctx, getInput)
	if err != nil {
		noSuchFn := new(types.NoSuchFunctionExists)
		if errors.As(err, &noSuchFn) {
			return noSuchFn
		}
		return fmt.Errorf("GetFunction: %w", err)
	}

	describeInput := &cloudfront.DescribeFunctionInput{
		Name: &i.functionName,
	}
	describeOut, err := i.client.DescribeFunction(ctx, describeInput)
	if err != nil {
		return fmt.Errorf("DescribeFunction: %w", err)
	}

	if _, err := i.functionStream.Write(getOut.FunctionCode); err != nil {
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
			Path: i.functionStream.FilePath,
		},
	}
	if err := writeFunctionToStream(fn, i.configStream); err != nil {
		return err
	}
	return nil
}

type WritableFile struct {
	io.Writer
	FilePath string
}
