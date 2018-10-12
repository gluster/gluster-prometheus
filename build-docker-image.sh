#! /bin/bash
set -e

# Allow overriding default docker command
DOCKER_CMD=${DOCKER_CMD:-docker}

VERSION="$(git describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2')"
BUILDDATE="$(date -u '+%Y-%m-%dT%H:%M:%S.%NZ')"
docker build \
        -t gluster/gluster-prometheus \
        --build-arg version="$VERSION" \
        --build-arg builddate="$BUILDDATE" \
        -f Dockerfile \
        .
