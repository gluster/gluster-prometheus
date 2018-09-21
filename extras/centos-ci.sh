#!/bin/bash

# This script will be called by the gluster_prometheus job script on centos-ci.
# This script sets up the centos-ci environment and runs the PR
# tests for gluster-prometheus.

# if anything fails, we'll abort
set -e

REQ_GO_VERSION='1.9.4'
# install Go
if ! yum -y install "golang >= $REQ_GO_VERSION"
then
	# not the right version, install manually
	# download URL comes from https://golang.org/dl/
	curl -O https://storage.googleapis.com/golang/go${REQ_GO_VERSION}.linux-amd64.tar.gz
	tar xzf go${REQ_GO_VERSION}.linux-amd64.tar.gz -C /usr/local
	export PATH=$PATH:/usr/local/go/bin
fi

# also needs git, hg, bzr, svn gcc and make
yum -y install git mercurial bzr subversion gcc make
yum -y install ShellCheck

export EXPORTER_SRC=$GOPATH/src/github.com/gluster/gluster-prometheus
cd "$EXPORTER_SRC"

# install the build and test requirements
./scripts/install-reqs.sh

# install vendored dependencies
make vendor-install

# verify build
make gluster-exporter

# run tests
make test TESTOPTIONS=-v
