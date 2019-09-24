#!bash

for spec in "linux amd64" "windows amd64 .exe" "darwin amd64"
do
    export GOOS=$(echo $spec | cut -d ' ' -f 1)
    export GOARCH=$(echo $spec | cut -d ' ' -f 2)
    EXTENSION=$(echo $spec | cut -d ' ' -f 3)
    PACKAGE="${GOOS}-${GOARCH}"
    echo ${PACKAGE}
    mkdir -p targets/${PACKAGE}
    go build -o "targets/${PACKAGE}/git-docs${EXTENSION}"
done