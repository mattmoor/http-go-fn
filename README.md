# http-go-fn

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/mattmoor/http-go-fn)
[![Go Report Card](https://goreportcard.com/badge/mattmoor/http-go-fn)](https://goreportcard.com/report/mattmoor/http-go-fn)
[![Releases](https://img.shields.io/github/release-pre/mattmoor/http-go-fn.svg?sort=semver)](https://github.com/mattmoor/http-go-fn/releases)
[![LICENSE](https://img.shields.io/github/license/mattmoor/http-go-fn.svg)](https://github.com/mattmoor/http-go-fn/blob/master/LICENSE)
[![codecov](https://codecov.io/gh/mattmoor/http-go-fn/branch/master/graph/badge.svg)](https://codecov.io/gh/mattmoor/http-go-fn)

This repository implements a Go function buildpack for wrapping functions matching `http.HandlerFunc`.
This buildpack is not standalone, it should be composed with the Paketo Go buildpacks.


# Build this buildpack

This buildpack can be built (from the root of the repo) with:

```shell
pack package-buildpack my-buildpack --config ./package.toml
```


# Use this buildpack

```shell
# This runs the http-go-fn buildpack at HEAD within the Paketo Go order.
# You can pin to a release by replacing ":main" below with a release tag
# e.g. ":v0.0.1"
pack build -v test-container \
  --pull-policy if-not-present \
  --buildpack gcr.io/paketo-buildpacks/go-dist:0.2.5 \
  --buildpack ghcr.io/mattmoor/http-go-fn:main \
  --buildpack gcr.io/paketo-buildpacks/go-mod-vendor:0.0.169 \
  --buildpack gcr.io/paketo-buildpacks/go-build:0.1.2
```


# Sample function

With this buildpack, users can define a trivial Go function that implements
[`http.HandlerFunc`](https://godoc.org/net/http#HandlerFunc).  For example,
the following function:

```go
package fn

import (
       "fmt"
       "net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
     fmt.Fprintf(w, "Hello World, %#v", r)
}
```


# Configuration

You can configure both the package containing the function and the name of
the function via the following configuration options in `project.toml`:

```toml
[[build.env]]
name = "HTTP_GO_PACKAGE"
value = "./blah"

[[build.env]]
name = "HTTP_GO_FUNCTION"
value = "MyCustomHandlerName"
```
