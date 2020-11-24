package function

import "github.com/paketo-buildpacks/packit"

type Detector struct {
	Package  string `envconfig:"HTTP_GO_PACKAGE" default:"."`
	Function string `envconfig:"HTTP_GO_FUNCTION" default:"Handler"`
}

func (d *Detector) Detect(dctx packit.DetectContext) (packit.DetectResult, error) {
	// TODO(mattmoor): Inline Ville detect logic.
	pkg := d.Package
	fn := d.Function

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
