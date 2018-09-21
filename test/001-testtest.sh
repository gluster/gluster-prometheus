#!/bin/bash

# This "test" exists merely to exercise the test "framework"

if [ "$EXPORTER_TEST_TEST" ]; then
	exit "$EXPORTER_TEST_TEST"
fi
exit 0
