#!/bin/bash

rm -rf buildpacks/bin/
mkdir -p buildpacks/bin/
CGO_ENABLED=0 go build -mod=vendor -o buildpacks/bin/run ./cmd/run
pushd buildpacks/bin
ln -s -r run detect
ln -s -r run build
popd
pack package-buildpack "${@}"
