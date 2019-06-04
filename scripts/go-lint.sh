#!/bin/bash

set -eo pipefail

echo "Lint golang sources..."

golangci-lint --config=scripts/golangci.yml run ./... -v || exit $?
echo
echo "ALL OK."
