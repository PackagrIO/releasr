#!/usr/bin/env bash


# because packagr/releasr depends on compiled binaries (even for testing) we'll be building the test binaries first in the "build container" and then
# executing them in a "runtime container" to get coverage/profiling data.
#
# this script executes the test binaries in the "runtime container"

set -e

echo "Args: $@"
echo  "/coverage/coverage-${1}.txt"

mkdir -p /coverage
echo "" > "/coverage/coverage-${1}.txt"

for d in $(find . -type f -name "test_binary_*"); do
    echo "Found TEST BINARY: ${d}"
        pushd $(dirname "$d")

        eval "./test_binary_${1} -test.coverprofile=profile.out"
        if [ -f profile.out ]; then
            cat profile.out >> "/coverage/coverage-${1}.txt"
            rm profile.out
        fi
        popd
done
