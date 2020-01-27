#!/bin/bash

for spec in "linux amd64" "windows amd64 .exe" "darwin amd64"
do
    export GOOS=$(echo $spec | cut -d ' ' -f 1)
    export GOARCH=$(echo $spec | cut -d ' ' -f 2)
    EXTENSION=$(echo $spec | cut -d ' ' -f 3)
    PACKAGE="${GOOS}-${GOARCH}"
    echo ${PACKAGE}
    mkdir -p git-docs-releases/${PACKAGE}
    go build -o "git-docs-releases/${PACKAGE}/git-docs${EXTENSION}" git-docs
done