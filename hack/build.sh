#!/bin/bash

mkdir -p buildpacks/bin/ || true
CGO_ENABLED=0 go build -mod=vendor -o buildpacks/bin/build ./cmd/run
CGO_ENABLED=0 go build -mod=vendor -o buildpacks/bin/detect ./vendor/github.com/vaikas/buildpackstuffhttp/cmd/detect
pack package-buildpack "${@}"
