#!/bin/bash

# Checks if the tools required for properly building, verifying and testing are installed.

TOOLS=(dep gometalinter)

MISSING=0

for tool in "${TOOLS[@]}"; do
  
  if ! type "$tool" >/dev/null 2>&1; then
    echo "$tool package is missing on the system"
    MISSING=1
  else
    echo "$tool package is available"
  fi
done

if [ $MISSING -ne 0 ]; then
  exit 1
fi
