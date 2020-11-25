package function

import (
	"testing"

	"github.com/paketo-buildpacks/packit"
)

func TestDetect(t *testing.T) {
	const goodWD = "../../" // where our go.mod file lives

	tests := []struct {
		name  string
		wd    string
		pkg   string
		fn    string
		match bool
	}{{
		name:  "default function",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/default",
		fn:    "Handler",
		match: true,
	}, {
		name:  "non-default function (no override)",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/nondefault",
		fn:    "Handler",
		match: false,
	}, {
		name:  "non-default function (correct override)",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/nondefault",
		fn:    "MyCustomHandler",
		match: true,
	}, {
		name:  "bad signature function",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata/nomatch",
		fn:    "Handler",
		match: false,
	}, {
		name:  "no functions",
		wd:    goodWD,
		pkg:   "./pkg/function/testdata",
		fn:    "Handler",
		match: false,
	}, {
		name:  "no go.mod",
		wd:    ".",
		pkg:   "./testdata/default",
		fn:    "Handler",
		match: false,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Detector{
				Package:  test.pkg,
				Function: test.fn,
			}
			p, err := d.Detect(packit.DetectContext{
				WorkingDir: test.wd,
			})
			if err != nil && test.match {
				t.Fatal("Unexpected error:", err)
			} else if err == nil && !test.match {
				t.Fatal("Unexpected match:", p)
			}
		})
	}
}
