# http-go-fn

Playground experimenting with Go functions for `http.HandlerFunc`

# Build this buildpack

This buildpack can be built (from the root of the repo) with:


```shell
pack package-buildpack my-buildpack --config ./package.toml
```

# Use this buildpack

```shell
pack build blah --buildpack ghcr.io/mattmoor/http-go-fn:main
```

# Sample function

With this buildpack, users can define a trivial Go function that implements
[`http.HandlerFunc`](https://godoc.org/net/http#HandlerFunc).  For example,
the following function:

```go
package fn

import "net/http"

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
