.PHONY: build-image run-interactive

build-prepare:
	go get github.com/jteeuwen/go-bindata/...
	go get github.com/elazarl/go-bindata-assetfs/...

build:
	go build -o bin/mgit main.go

run-serve:
	go run main.go serve

build-docker:
	docker build . -t magic-git

run-serve-docker:
	docker run -it --rm -v "$(shell pwd)":/usr/src/myapp -w /usr/src/myapp --user $(shell id -u):$(shell id -g) magic-git serve
