package function

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"
)

func TestBuild(t *testing.T) {
	const (
		layersPath = "/layers"
		// These don't need to be real, we just need to be able to check them.
		pkg = "paketo.io/my-fn"
		fn  = "MyHandler"
	)
	wantBuildPlan := packit.BuildResult{
		Layers: []packit.Layer{{
			Name:  "http-go-funtion-cmd",
			Path:  filepath.Join(layersPath, "http-go-funtion-cmd"),
			Build: true,
			BuildEnv: packit.Environment{
				"BP_GO_TARGETS.override": targetPackage,
			},
		}},
	}
	buf := bytes.NewBuffer(nil)
	if err := mainTemplate.Execute(buf, struct {
		Package  string
		Function string
	}{
		Package:  pkg,
		Function: fn,
	}); err != nil {
		t.Fatal("mainTemplate.Execute() =", err)
	}
	wantFileContents := buf.String()

	tests := []struct {
		name    string
		plan    packit.BuildpackPlan
		success bool
	}{{
		name: "successful build",
		plan: packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{{
				Name: "unrelated-entry",
			}, {
				Name: "http-go-function",
				Metadata: map[string]interface{}{
					"package":  pkg,
					"function": fn,
				},
			}},
		},
		success: true,
	}, {
		name: "missing plan entry",
		plan: packit.BuildpackPlan{
			Entries: []packit.BuildpackPlanEntry{{
				Name: "unrelated-entry",
			}},
		},
		success: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := Builder{
				Logger: scribe.NewLogger(ioutil.Discard),
			}
			dir, err := ioutil.TempDir("", "")
			if err != nil {
				t.Fatal("TempDir() =", err)
			}
			defer os.RemoveAll(dir)

			bp, err := b.Build(packit.BuildContext{
				WorkingDir: dir,
				Layers: packit.Layers{
					Path: layersPath,
				},
				Plan: test.plan,
			})
			if err != nil && test.success {
				t.Fatal("Unexpected error:", err)
			} else if err == nil && !test.success {
				t.Fatal("Unexpected failure:", bp)
			}
			if err != nil {
				return
			}

			// Check that the build plan matches what we want.
			if !cmp.Equal(bp, wantBuildPlan) {
				t.Errorf("Build (-want, +got): %s", cmp.Diff(wantBuildPlan, bp))
			}

			gotFileContents, err := ioutil.ReadFile(filepath.Join(dir, targetPackage, "main.go"))
			if err != nil {
				t.Fatal("ReadFile() =", err)
			}

			if wantFileContents != string(gotFileContents) {
				t.Fatalf("ReadFile() = %q, wanted %q", string(gotFileContents), wantFileContents)
			}
		})
	}
}
