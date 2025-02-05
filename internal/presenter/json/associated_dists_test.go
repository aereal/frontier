package json_test

import (
	"bytes"
	"testing"

	"github.com/aereal/frontier"
	"github.com/aereal/frontier/internal/presenter/json"
	"github.com/google/go-cmp/cmp"
)

var (
	input = []frontier.FunctionAssociation{
		{
			EventType: "viewer-response",
			CacheBehavior: frontier.CacheBehavior{
				CachePolicyID:  "1234-5678",
				TargetOriginID: "test-origin",
				IsDefault:      true,
			},
			Distribution: frontier.AssociatedDistribution{
				DomainName: "dist.test",
				ARN:        "arn:aws:cloudfront::123456789012:distribution/0XDEADBEAF",
				ID:         "0XDEADBEAF",
				IsEnabled:  true,
				IsStaging:  false,
				Status:     "Deployed",
			},
			Function: frontier.AssociatedFunction{
				ARN: "arn:aws:cloudfront::123456789012:function/test-fn",
			},
		},
	}
)

func TestAssociatedDistributionsPresenter(t *testing.T) {
	testCases := []struct {
		name    string
		options []json.NewAssociatedDistributionsPresenterOption
		input   []frontier.FunctionAssociation
		want    string
	}{
		{
			name:  "compact",
			input: input,
			want:  `{"Distribution":{"DomainName":"dist.test","ARN":"arn:aws:cloudfront::123456789012:distribution/0XDEADBEAF","ID":"0XDEADBEAF","IsEnabled":true,"IsStaging":false,"Status":"Deployed"},"CacheBehavior":{"CachePolicyID":"1234-5678","TargetOriginID":"test-origin","IsDefault":true},"EventType":"viewer-response","Function":{"ARN":"arn:aws:cloudfront::123456789012:function/test-fn"}}` + "\n",
		},
		{
			name:  "pretty",
			input: input,
			options: []json.NewAssociatedDistributionsPresenterOption{
				json.Pretty(true),
			},
			want: "{\n  \"Distribution\": {\n    \"DomainName\": \"dist.test\",\n    \"ARN\": \"arn:aws:cloudfront::123456789012:distribution/0XDEADBEAF\",\n    \"ID\": \"0XDEADBEAF\",\n    \"IsEnabled\": true,\n    \"IsStaging\": false,\n    \"Status\": \"Deployed\"\n  },\n  \"CacheBehavior\": {\n    \"CachePolicyID\": \"1234-5678\",\n    \"TargetOriginID\": \"test-origin\",\n    \"IsDefault\": true\n  },\n  \"EventType\": \"viewer-response\",\n  \"Function\": {\n    \"ARN\": \"arn:aws:cloudfront::123456789012:function/test-fn\"\n  }\n}\n",
		},
		{
			name:  "empty",
			input: []frontier.FunctionAssociation{},
			want:  "",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out := new(bytes.Buffer)
			presenter := json.NewAssociatedDistributionsPresenter(out, tc.options...)
			presenter.PresentAssociatedDistributions(tc.input)
			got := out.String()
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Logf("got: %q", got)
				t.Errorf("(-want, +got):\n%s", diff)
			}
		})
	}
}
