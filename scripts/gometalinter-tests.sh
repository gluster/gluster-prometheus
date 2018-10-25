#!/bin/bash

set -eo pipefail

echo "Lint golang sources..."

gometalinter -j1 --sort=path --sort=line --sort=column \
             --enable="gofmt" \
	     --disable="gocyclo" \
	     --exclude="Subprocess launching should be audited" \
	     --exclude="Subprocess launched with variable" \
	     --deadline 9m --vendor --debug ./... |& stdbuf -oL awk '/linter took/ || !/^DEBUG/' || exit $?

echo
echo "ALL OK."
