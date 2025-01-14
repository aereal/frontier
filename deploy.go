//go:generate go run go.uber.org/mock/mockgen -package frontier_test -destination mock_cloudfront_client_test.go github.com/aereal/frontier CloudFrontClient

package frontier

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type CloudFrontClient interface {
	CreateFunction(ctx context.Context, params *cloudfront.CreateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.CreateFunctionOutput, error)
	GetFunction(ctx context.Context, params *cloudfront.GetFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.GetFunctionOutput, error)
	PublishFunction(ctx context.Context, params *cloudfront.PublishFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.PublishFunctionOutput, error)
	UpdateFunction(ctx context.Context, params *cloudfront.UpdateFunctionInput, optFns ...func(*cloudfront.Options)) (*cloudfront.UpdateFunctionOutput, error)
}

type Deployer struct {
	client CloudFrontClient
	logger *zap.Logger
}

func NewDeployer(client CloudFrontClient, logger *zap.Logger) *Deployer {
	d := &Deployer{client: client, logger: zap.NewNop()}
	if logger != nil {
		d.logger = logger
	}
	return d
}

func (d *Deployer) Deploy(ctx context.Context, configPath string, publish bool) error {
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	fn := new(Function)
	if err := dec.Decode(fn); err != nil {
		return err
	}

	var etag *string
	existing, err := d.client.GetFunction(ctx, &cloudfront.GetFunctionInput{Name: &fn.Name})
	if err != nil {
		var notFoundErr *types.NoSuchFunctionExists
		if !errors.As(err, &notFoundErr) {
			return err
		}
		input, err := fn.toCreateInput()
		if err != nil {
			return err
		}
		out, err := d.client.CreateFunction(ctx, input)
		if err != nil {
			return err
		}
		etag = out.ETag
	} else {
		input, err := fn.toUpdateInput(existing.ETag)
		if err != nil {
			return err
		}
		out, err := d.client.UpdateFunction(ctx, input)
		if err != nil {
			return err
		}
		etag = out.ETag
	}

	if publish && etag != nil {
		input := &cloudfront.PublishFunctionInput{
			Name:    &fn.Name,
			IfMatch: etag,
		}
		if _, err := d.client.PublishFunction(ctx, input); err != nil {
			return err
		}
	}
	return nil
}
