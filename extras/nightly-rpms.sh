#!/bin/bash

# This scripts builds RPMs from the current git head.
# The script needs be run from the root of the repository
# NOTE: RPMs are built only for EL7 (CentOS7) distributions.

set -e

##
## Set up build environment
##
RESULTDIR=${RESULTDIR:-$PWD/rpms}
BUILDDIR=$PWD/$(mktemp -d nightlyrpmXXXXXX)

BASEDIR=$(dirname "$0")
GPROMCLONE=$(realpath "$BASEDIR/..")

yum -y install make mock rpm-build golang

export GOPATH=$BUILDDIR/go
mkdir -p "$GOPATH"/{bin,pkg,src}
export PATH=$GOPATH/bin:$PATH

GPROMSRC=$GOPATH/src/github.com/gluster/gluster-prometheus
mkdir -p "$GOPATH/src/github.com/gluster"
ln -s "$GPROMCLONE" "$GPROMSRC"

"$GPROMSRC/scripts/install-reqs.sh"

##
## Prepare gluster-prometheus archives and specfile for building RPMs
##
pushd "$GPROMSRC"

FULL_VERSION=$(./scripts/pkg-version --full)

# Create a vendored dist archive
DISTDIR=$BUILDDIR SIGN=no make dist-vendor

# Copy over specfile to the BUILDDIR and modify it to use the current Git HEAD versions
cp ./extras/rpms/* "$BUILDDIR"

popd #GPROMSRC

pushd "$BUILDDIR"

DISTARCHIVE="gluster-exporter-$FULL_VERSION-vendor.tar.xz"
SPEC=gluster-exporter.spec

# Create SRPM
mkdir -p rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
cp "$BUILDDIR/$DISTARCHIVE" rpmbuild/SOURCES
cp $SPEC rpmbuild/SPECS
SRPM=$(rpmbuild --define "_topdir $PWD/rpmbuild" -bs rpmbuild/SPECS/$SPEC | cut -d\  -f2)

# Build RPM from SRPM using mock
mkdir -p "$RESULTDIR"
/usr/bin/mock -r epel-7-x86_64 --resultdir="$RESULTDIR" --rebuild "$SRPM"

popd #BUILDDIR

## Cleanup
rm -rf "$BUILDDIR"
