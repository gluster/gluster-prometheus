#!/bin/bash

TAR="tar"
GOPATH=$(go env GOPATH)
GOBINDIR=$GOPATH/bin
INSTALL_GOMETALINTER=${INSTALL_GOMETALINTER:-yes}


install_tools_darwin() {
  if type brew >/dev/null 2>&1; then
    echo "brew is already installed"
  else
    echo "Installing brew."
    mkdir homebrew && curl -L https://github.com/Homebrew/brew/tarball/master | tar xz --strip 1 -C homebrew
  fi

  if type gtar >/dev/null 2>&1; then
    echo "gnu-tar (gtar) is already installed"
  else
    echo "Installing gnu-tar."
    brew install gnu-tar
  fi

  TAR="gtar"
}

bootstrap_platform() {
  case "$OSTYPE" in
    solaris*) echo "SOLARIS" ;;
    darwin*)  echo "OSX" ; install_tools_darwin ;;
    linux*)   echo "LINUX" ;;
    bsd*)   echo "BSD" ;;
    msys*)  echo "WINDOWS" ;;
    *)  echo "unknown: $OSTYPE" ;;
  esac
}

install_dep() {
  DEPVER="v0.5.0"
  DEPURL="https://github.com/golang/dep/releases/download/${DEPVER}/dep-linux-amd64"
  if type dep >/dev/null 2>&1; then
    local version
    version=$(dep version | awk '/^ version/{print $3}')
    if [[ $version == "$DEPVER" || $version >  $DEPVER ]]; then
      echo "dep ${DEPVER} or greater is already installed"
      return
    fi
  fi

  echo "Installing dep. Version: ${DEPVER}"
  DEPBIN=$GOPATH/bin/dep
  curl -L -o "$DEPBIN" $DEPURL
  chmod +x "$DEPBIN"
}

install_gometalinter() {
  LINTER_VER="3.0.0"
  LINTER_TARBALL="gometalinter-${LINTER_VER}-linux-amd64.tar.gz"
  LINTER_URL="https://github.com/alecthomas/gometalinter/releases/download/v${LINTER_VER}/${LINTER_TARBALL}"

  if type gometalinter >/dev/null 2>&1; then
    echo "gometalinter already installed"
    return
  fi

  echo "Installing gometalinter. Version: ${LINTER_VER}"
  curl -L -o "$GOBINDIR/$LINTER_TARBALL" $LINTER_URL
  $TAR -zxf "$GOBINDIR/$LINTER_TARBALL" --overwrite --strip-components 1 --exclude={COPYING,*.md} -C "$GOBINDIR"
  rm -f "$GOBINDIR/$LINTER_TARBALL"
}

bootstrap_platform
install_dep
if [ "$INSTALL_GOMETALINTER" == "yes" ];then
    install_gometalinter
fi
