[![status][ci-status-badge]][ci-status]
[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev]

# frontier

frontier is a deployment tool for [Amazon CloudFront Functions][cf-functions].

The concept is heavily inspired by [Lambroll][].

## Installation

```sh
go install github.com/aereal/frontier@latest
```

## License

See LICENSE file.

[pkg-go-dev]: https://pkg.go.dev/github.com/aereal/frontier
[pkg-go-dev-badge]: https://pkg.go.dev/badge/aereal/frontier
[ci-status-badge]: https://github.com/aereal/frontier/workflows/CI/badge.svg?branch=main
[ci-status]: https://github.com/aereal/frontier/actions/workflows/CI
[cf-functions]: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/cloudfront-functions.html
[lambroll]: https://github.com/fujiwara/lambroll
