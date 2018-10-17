#!/bin/bash

set -o pipefail

echo "Lint golang sources..."

if gometalinter -j1 --sort=path --sort=line --sort=column \
                --enable="gofmt" \
	        --disable="gocyclo" \
	        --exclude="Subprocess launching should be audited" \
	        --exclude="Subprocess launched with variable" \
	        --deadline 9m --vendor --debug ./... |& stdbuf -oL awk '/linter took/ || !/^DEBUG/'; then
    echo
    echo "ALL OK."
fi
