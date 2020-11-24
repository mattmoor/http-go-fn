package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/paketo-buildpacks/packit"

	"github.com/mattmoor/http-go-fn/pkg/function"
)

func main() {
	d := function.Detector{}
	if err := envconfig.Process("", &d); err != nil {
		log.Fatal(err)
	}

	b := function.Builder{}
	if err := envconfig.Process("", &b); err != nil {
		log.Fatal(err)
	}

	packit.Run(
		d.Detect,
		b.Build,
	)
}
