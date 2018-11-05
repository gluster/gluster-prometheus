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

INSTALL_GOMETALINTER=no "$GPROMSRC/scripts/install-reqs.sh"

##
## Prepare gluster-prometheus archives and specfile for building RPMs
##
pushd "$GPROMSRC"

VERSION=$(./scripts/pkg-version --version)
RELEASE=$(./scripts/pkg-version --release)
FULL_VERSION=$(./scripts/pkg-version --full)

# Create a vendored dist archive
DISTDIR=$BUILDDIR SIGN=no make dist-vendor

# Copy over specfile to the BUILDDIR and modify it to use the current Git HEAD versions
cp ./extras/rpms/* "$BUILDDIR"

popd #GPROMSRC

pushd "$BUILDDIR"

DISTARCHIVE="gluster-prometheus-exporter-$FULL_VERSION-vendor.tar.xz"
SPEC=gluster-prometheus-exporter.spec
sed -i -E "
# Use bundled always
s/with_bundled 0/with_bundled 1/;
# Replace version with HEAD version
s/%global gluster_prom_ver[[:space:]]+(.+)$/%global gluster_prom_ver $VERSION/;
# Replace release with proper release
s/%global gluster_prom_rel[[:space:]]+(.+)$/%global gluster_prom_rel $RELEASE/;
# Replace Source0 with generated archive
s/^Source0:[[:space:]]+.*.tar.xz/Source0: $DISTARCHIVE/;
" $SPEC

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
