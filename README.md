[![status][ci-status-badge]][ci-status]
[![PkgGoDev][pkg-go-dev-badge]][pkg-go-dev]

# frontier

frontier is a deployment tool for [Amazon CloudFront Functions][cf-functions].

The concept is heavily inspired by [Lambroll][].

## Synopsis

Simply run as below:

```
frontier deploy
```

`frontier deploy` **does**:

- create or update function
- publish the function by default
  - you can stop this behavior by using `--publish=false`

`frontier deploy` _does not_:

- compile your function code implicitly
- or anything else

### Function Config (function.yml)

The function config is almost same as `CreateFunction` or `UpdateFunction`'s input except of `Code`.

```yaml
---

name: your-edge-function
config:
  comment: this is edge function
  runtime: cloudfront-js-1.0
code:
  path: ./path/to/fn.js
```

fn.js:

```javascript
function handler(event) {
  var response = event.response,
    origin = event.request.headers.origin;
  response.headers["access-control-allow-origin"] = {
    value: origin.value,
  };
  return response;
}
```

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
