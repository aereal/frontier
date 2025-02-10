package listdist_test

import (
	"regexp"
	"testing"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/controller/listdist"
)

func TestListDistributionsCriteria(t *testing.T) {
	domainNamePattern, err := regexp.Compile(`[.]test$`) //nolint:gocritic
	if err != nil {
		panic(err)
	}
	association := frontier.FunctionAssociation{
		EventType: "viewer-request",
		Distribution: frontier.AssociatedDistribution{
			IsEnabled:  true,
			DomainName: "dist.test",
		},
		Function: frontier.AssociatedFunction{
			ARN: "arn:aws:cloudfront::123456789012:function/test-fn",
		},
	}
	testCases := []struct {
		name        string
		criteria    *listdist.Criteria
		association frontier.FunctionAssociation
		want        bool
	}{
		{name: "empty", criteria: listdist.NewCriteria(), want: true},
		{
			name: "single criterion, matched association",
			criteria: listdist.NewCriteria(
				listdist.EqualEventType("viewer-request"),
			),
			association: association,
			want:        true,
		},
		{
			name: "single criterion, MISMATCHED association",
			criteria: listdist.NewCriteria(
				listdist.EqualEventType("viewer-response"),
			),
			association: association,
			want:        false,
		},
		{
			name: "multiple criteria, satisfied all",
			criteria: listdist.NewCriteria(
				listdist.EqualEventType("viewer-request"),
				listdist.EqualDistributionIsEnabled(true),
				listdist.EqualDistributionDomainName("dist.test"),
			),
			association: association,
			want:        true,
		},
		{
			name: "multiple criteria, NOT satisfied all",
			criteria: listdist.NewCriteria(
				listdist.EqualEventType("viewer-request"),
				listdist.EqualDistributionIsEnabled(false),
			),
			association: association,
			want:        false,
		},
		{
			name: "function ARN",
			criteria: listdist.NewCriteria(
				listdist.EqualFunctionArn("arn:aws:cloudfront::123456789012:function/test-fn"),
			),
			association: association,
			want:        true,
		},
		{
			name: "domain name pattern",
			criteria: listdist.NewCriteria(
				listdist.MatchDistributionDomainName(domainNamePattern),
			),
			association: association,
			want:        true,
		},
		{
			name: "multiple domain name criteria passed",
			criteria: listdist.NewCriteria(
				listdist.EqualDistributionDomainName("another.test"), // not matched
				listdist.MatchDistributionDomainName(domainNamePattern),
			),
			association: association,
			want:        true,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.criteria.Satisfy(tc.association)
			if got != tc.want {
				t.Errorf("result mismatch:\n\twant: %v\n\t got: %v", tc.want, got)
			}
		})
	}
}
