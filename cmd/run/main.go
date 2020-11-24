package main

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/scribe"

	"github.com/mattmoor/http-go-fn/pkg/function"
)

func main() {
	logger := scribe.NewLogger(os.Stdout)

	d := function.Detector{}
	if err := envconfig.Process("", &d); err != nil {
		log.Fatal(err)
	}

	b := function.Builder{
		Logger: logger,
	}
	if err := envconfig.Process("", &b); err != nil {
		log.Fatal(err)
	}

	packit.Run(
		d.Detect,
		b.Build,
	)
}
