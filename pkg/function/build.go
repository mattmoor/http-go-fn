package function

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

// Builder is a stateful implementation of packit.BuildFunc to implement
// the build phase of a Paketo buildpack.
type Builder struct {
	// Logger is used to emit waypoints through the build phase of th elifecycle.
	Logger scribe.Logger
}

const targetPackage = "./http-cmd/function"

// Build is a member function that implements packit.BuildFunc
func (b *Builder) Build(bctx packit.BuildContext) (packit.BuildResult, error) {
	b.Logger.Title("%s %s", bctx.BuildpackInfo.Name, bctx.BuildpackInfo.Version)
	defer b.Logger.Break()

	pkg, fn, err := b.getPkgFn(bctx)
	if err != nil {
		return packit.BuildResult{}, err
	}
	b.Logger.Process("Package:  %s", pkg)
	b.Logger.Process("Function: %s", fn)

	if err := os.MkdirAll(filepath.Join(bctx.WorkingDir, targetPackage), os.ModePerm); err != nil {
		return packit.BuildResult{}, err
	}

	p := filepath.Join(bctx.WorkingDir, targetPackage, "main.go")
	mg, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return packit.BuildResult{}, err
	}
	defer mg.Close()

	if err := mainTemplate.Execute(mg, struct {
		Package  string
		Function string
	}{
		Package:  pkg,
		Function: fn,
	}); err != nil {
		return packit.BuildResult{}, err
	}

	return packit.BuildResult{
		Layers: []packit.Layer{{
			Name:  "http-go-funtion-cmd",
			Path:  filepath.Join(bctx.Layers.Path, "http-go-funtion-cmd"),
			Build: true,
			BuildEnv: packit.Environment{
				"BP_GO_TARGETS.override": targetPackage,
			},
		}},
	}, nil
}

func (b *Builder) getPkgFn(bctx packit.BuildContext) (string, string, error) {
	for _, entry := range bctx.Plan.Entries {
		if entry.Name != "http-go-function" {
			continue
		}
		return entry.Metadata["package"].(string), entry.Metadata["function"].(string), nil
	}

	return "", "", errors.New("missing metadata for http-go-function")
}
