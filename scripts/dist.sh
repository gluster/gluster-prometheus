#!/usr/bin/env bash

# This script builds a dist tarball of the source
# This should only be called from the root of the repo

VENDOR=${VENDOR:-no}
OUTDIR=${DISTDIR:-.}
SIGN=${SIGN:-yes}

VERSION=$("$(dirname "$0")/pkg-version" --full)

BASENAME=gluster-prometheus
SUFFIX="exporter-$VERSION"
case $VENDOR in
  yes|y|Y)
    SUFFIX="exporter-$VERSION-vendor"
    ;;
esac
TARNAME="${BASENAME}-${SUFFIX}"

TARFILE=$OUTDIR/$TARNAME.tar
ARCHIVE=$TARFILE.xz
SIGNFILE=$ARCHIVE.asc

# Cleanup old archives
if [[ -f $ARCHIVE ]]; then
  rm "$ARCHIVE"
fi
if [[ -f $SIGNFILE ]]; then
  rm "$SIGNFILE"
fi

echo "Creating dist archive $ARCHIVE"
git archive -o "$TARFILE" --prefix "$BASENAME/" HEAD
tar --transform "s/^\\./$BASENAME/" -rf "$TARFILE" ./VERSION || exit 1
tar --transform "s/^\\./$BASENAME/" -rf "$TARFILE" ./GIT_SHA_FULL || exit 1
case $VENDOR in
  yes|y|Y)
    tar --transform "s/^\\./$BASENAME/" -rf "$TARFILE" ./vendor || exit 1
    ;;
esac

xz "$TARFILE" || exit 1
echo "Created dist archive $ARCHIVE"


# Sign the generated archive
case $SIGN in
  yes|y|Y)
    echo "Signing dist archive"
    gpg --armor --output "$SIGNFILE" --detach-sign "$ARCHIVE" || exit 1
    echo "Signed dist archive, signature in $SIGNFILE"
    ;;
esac

# Remove the VERSION file, it is no longer needed and would harm normal builds
rm VERSION

