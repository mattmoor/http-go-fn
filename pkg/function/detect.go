package function

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/vaikas/gofunctypechecker/pkg/detect"
)

type Detector struct {
	Package  string `envconfig:"HTTP_GO_PACKAGE" default:"."`
	Function string `envconfig:"HTTP_GO_FUNCTION" default:"Handler"`
}

// Valid function signature for HTTP Handler is:
// func(http.ResponseWriter, *http.Request)
var (
	validFunctions = []detect.FunctionSignature{{
		In: []detect.FunctionArg{{
			ImportPath: "net/http",
			Name:       "ResponseWriter",
		}, {
			ImportPath: "net/http",
			Name:       "Request",
			Pointer:    true,
		}},
	}}
	detector = detect.NewDetector(validFunctions)
)

func (d *Detector) Detect(dctx packit.DetectContext) (packit.DetectResult, error) {
	moduleName, err := readModuleName(dctx)
	if err != nil {
		return packit.DetectResult{}, err
	}
	pkg := filepath.Join(moduleName, d.Package)
	fn := d.Function

	if err := d.checkFunction(dctx, pkg, fn); err != nil {
		return packit.DetectResult{}, err
	}

	return packit.DetectResult{
		Plan: packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{{
				Name: "http-go-function",
			}},
			Requires: []packit.BuildPlanRequirement{{
				Name: "http-go-function",
				Metadata: map[string]interface{}{
					"package":  pkg,
					"function": fn,
				},
			}},
		},
	}, nil
}

func (d *Detector) checkFunction(dctx packit.DetectContext, pkg, fn string) error {
	// read all go files from the directory that was given. Note that if no directory (HTTP_GO_PACKAGE)
	// was given, this is ./
	files, err := filepath.Glob(filepath.Join(dctx.WorkingDir, d.Package, "*.go"))
	if err != nil {
		return err
	}

	for _, f := range files {
		srcbuf, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		deets, err := detector.CheckFile(&detect.Function{
			File:   f,
			Source: string(srcbuf),
		})
		if err != nil {
			return err
		}
		if deets == nil {
			continue
		}

		if deets.Name != fn {
			// TODO(mattmoor): Add help text to tell the user how to properly configure project.toml,
			// or make the defaulting smart when it is not explicitly specified.
			log.Printf("Found supported function %q in package %q signature %q", deets.Name, deets.Package, deets.Signature)
			continue
		}
		return nil
	}

	return fmt.Errorf("Unable to find function %q in %q with matching signature.", fn, pkg)
}

// readModuleName is a terrible hack for yanking the module from go.mod file.
// Should be replaced with something that actually understands go...
func readModuleName(dctx packit.DetectContext) (string, error) {
	modFile, err := os.Open(filepath.Join(dctx.WorkingDir, "go.mod"))
	if err != nil {
		return "", err
	}
	defer modFile.Close()

	scanner := bufio.NewScanner(modFile)
	for scanner.Scan() {
		pieces := strings.Split(scanner.Text(), " ")
		if len(pieces) >= 2 && pieces[0] == "module" {
			return pieces[1], nil
		}
	}
	return "", nil
}
