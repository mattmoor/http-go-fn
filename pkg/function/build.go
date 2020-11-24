package function

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit"
)

type Builder struct {
	// TODO(mattmoor): envconfig
}

const targetPackage = "./http-cmd/function"

func (b *Builder) Build(bctx packit.BuildContext) (packit.BuildResult, error) {
	pkg, fn, err := b.getPkgFn(bctx)
	if err != nil {
		return packit.BuildResult{}, err
	}
	log.Print("Package: ", pkg)
	log.Print("Function: ", fn)

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

	return "", "", errors.New("Missing metadata for http-go-function")
}
