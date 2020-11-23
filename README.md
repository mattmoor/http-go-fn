# hello-buildpacks

This is an adaptation of the [CNCF buildpacks samples repo](https://github.com/buildpacks/samples)
to create a Github "template" repository for creating new buildpacks.

# Build this buildpack

This buildpack can be built (from the root of the repo) with:

```shell
pack package-buildpack my-buildpack --config ./package.toml
```

# Use this buildpack

```shell
pack build blah --buildpack my-buildpack
```

# Sample output

```
===> DETECTING
hello-buildpacks 0.0.1
===> ANALYZING
Previous image with name "blah" not found
===> RESTORING
===> BUILDING
---> Buildpack Template
     platform_dir files:
       /platform:
       total 12
       drwxr-xr-x 1 root root 4096 Jan  1  1980 .
       drwxr-xr-x 1 root root 4096 Nov 14 15:50 ..
       drwxr-xr-x 1 root root 4096 Jan  1  1980 env
       
       /platform/env:
       total 8
       drwxr-xr-x 1 root root 4096 Jan  1  1980 .
       drwxr-xr-x 1 root root 4096 Jan  1  1980 ..
     env_dir: /platform/env
     env vars:
       declare -x CNB_BUILDPACK_DIR="/cnb/buildpacks/hello-buildpacks/0.0.1"
       declare -x CNB_STACK_ID="io.buildpacks.stacks.bionic"
       declare -x HOME="/home/cnb"
       declare -x HOSTNAME="e81e3627481f"
       declare -x OLDPWD
       declare -x PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
       declare -x PWD="/workspace"
       declare -x SHLVL="1"
     layers_dir: /layers/hello-buildpacks
     plan_path: /tmp/plan.545894262/hello-buildpacks/plan.toml
     plan contents:
       [[entries]]
         name = "some-thing"

---> Done
===> EXPORTING
Adding 1/1 app layer(s)
Adding layer 'launcher'
Adding layer 'config'
Adding label 'io.buildpacks.lifecycle.metadata'
Adding label 'io.buildpacks.build.metadata'
Adding label 'io.buildpacks.project.metadata'
Warning: default process type 'web' not present in list []
*** Images (8e56835297b3):
      blah
Successfully built image blah
```