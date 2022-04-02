# GCP Artifact Registry Plugin

This plugin allows uploading artifacts to artifact registry.

## Steps

1. Run `make` at project root to build protobufs, and binaries for all platforms

## Building with Docker

To build plugins for release you can use the `build-docker` Makefile target, this will 
build your plugin for all architectures and create zipped artifacts which can be uploaded
to an artifact manager such as GitHub releases.

The built artifacts will be output in the `./releases` folder.

```shell
make build-docker

rm -rf ./releases
DOCKER_BUILDKIT=1 docker build --output releases --progress=plain .
#1 [internal] load .dockerignore
#1 transferring context: 2B done
#1 DONE 0.0s

#...

#14 [export_stage 1/1] COPY --from=build /go/plugin/bin/*.zip .
#14 DONE 0.1s

#15 exporting to client
#15 copying files 36.45MB 0.1s done
#15 DONE 0.1s
```

## Building and releasing with GitHub Actions

When cloning the template a default GitHub Action is created at the path `.github/workflows/build-plugin.yaml`. You can use this action to automatically build and release your plugin.

The action has two main phases:
1. **Build** - This phase builds the plugin binaries for all the supported architectures. It is triggered when pushing
   to a branch or on pull requests.
1. **Release** - This phase creates a new GitHub release containing the built plugin. It is triggered when pushing tags
   which starting with `v`, for example `v0.1.0`.

You can enable this action by clicking on the `Actions` tab in your GitHub repository and enabling GitHub Actions.
