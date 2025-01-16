package frontier

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

type Deployer struct {
	client CFForDeploy
}

func NewDeployer(client CFForDeploy) *Deployer {
	d := &Deployer{client: client}
	return d
}

func (d *Deployer) Deploy(ctx context.Context, configPath string, publish bool) error {
	fn, err := parseConfigFromPath(configPath)
	if err != nil {
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
