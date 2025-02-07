package listdist

import (
	"context"
	"io"
	"iter"
	"slices"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/internal/cf"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
)

func NewController(clientProvider cf.Provider) *Controller {
	return &Controller{
		clientProvider: clientProvider,
	}
}

type Controller struct {
	clientProvider cf.Provider
}

func (c *Controller) ListDistributions(ctx context.Context, output io.Writer, criteria *Criteria) ([]frontier.FunctionAssociation, error) {
	client, err := c.clientProvider.ProvideCloudFrontClient(ctx)
	if err != nil {
		return nil, err
	}
	var fnAssociations []frontier.FunctionAssociation
	paginator := cloudfront.NewListDistributionsPaginator(client, &cloudfront.ListDistributionsInput{})
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, dist := range out.DistributionList.Items {
			fnAssociations = slices.AppendSeq(fnAssociations, criteria.filtered(iterateOverDist(dist)))
		}
	}
	return fnAssociations, nil
}

func convertSDKFunctionAssociations(in types.FunctionAssociation, dist types.DistributionSummary, isDefault bool, cachePolicyID string, targetOriginID *string) frontier.FunctionAssociation {
	var ret frontier.FunctionAssociation
	ret.EventType = string(in.EventType)
	ret.Function.ARN = *in.FunctionARN
	if dist.ARN != nil {
		ret.Distribution.ARN = *dist.ARN
	}
	if dist.DomainName != nil {
		ret.Distribution.DomainName = *dist.DomainName
	}
	if dist.Id != nil {
		ret.Distribution.ID = *dist.Id
	}
	if dist.Enabled != nil {
		ret.Distribution.IsEnabled = *dist.Enabled
	}
	if dist.Staging != nil {
		ret.Distribution.IsStaging = *dist.Staging
	}
	if dist.Status != nil {
		ret.Distribution.Status = *dist.Status
	}
	ret.CacheBehavior.CachePolicyID = cachePolicyID
	ret.CacheBehavior.IsDefault = isDefault
	if targetOriginID != nil {
		ret.CacheBehavior.TargetOriginID = *targetOriginID
	}
	return ret
}

func iterateOverDist(dist types.DistributionSummary) iter.Seq[frontier.FunctionAssociation] {
	return func(yield func(frontier.FunctionAssociation) bool) {
		for association := range iterateDefaultCacheBehavior(dist, dist.DefaultCacheBehavior) {
			if !yield(association) {
				return
			}
		}
		for association := range iterateCustomCacheBehavior(dist, dist.CacheBehaviors) {
			if !yield(association) {
				return
			}
		}
	}
}

func iterateCustomCacheBehavior(dist types.DistributionSummary, cbs *types.CacheBehaviors) iter.Seq[frontier.FunctionAssociation] {
	return func(yield func(frontier.FunctionAssociation) bool) {
		if cbs == nil {
			return
		}
		for _, cb := range cbs.Items {
			if cb.FunctionAssociations == nil {
				continue
			}
			for _, association := range cb.FunctionAssociations.Items {
				if !yield(convertSDKFunctionAssociations(association, dist, false, dereference(cb.CachePolicyId), cb.TargetOriginId)) {
					return
				}
			}
		}
	}
}

func iterateDefaultCacheBehavior(dist types.DistributionSummary, cb *types.DefaultCacheBehavior) iter.Seq[frontier.FunctionAssociation] {
	return func(yield func(frontier.FunctionAssociation) bool) {
		if cb == nil || cb.FunctionAssociations == nil {
			return
		}
		for _, association := range cb.FunctionAssociations.Items {
			if !yield(convertSDKFunctionAssociations(association, dist, true, dereference(cb.CachePolicyId), cb.TargetOriginId)) {
				return
			}
		}
	}
}

func dereference[T any](p *T) (ret T) { //nolint:ireturn
	if p != nil {
		ret = *p
	}
	return
}
