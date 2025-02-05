//go:generate go run go.uber.org/mock/mockgen -build_constraint !live -typed -write_command_comment=false -write_package_comment=false -write_source_comment=false -package cfmock -destination ./mock_gen.go github.com/aereal/frontier/internal/cf CloudFrontClient

package cfmock
