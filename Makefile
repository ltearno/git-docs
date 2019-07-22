all: build

.PHONY: build-prepare
build-prepare:
	go get -u github.com/jteeuwen/go-bindata/...
	go get github.com/julienschmidt/httprouter

.PHONY: build-embed-assets
build-embed-assets:
	go-bindata -o assetsgen/assets.go -pkg assetsgen assets/...

.PHONY: build
build: build-embed-assets
	go install git-docs

.PHONY: run-serve
run-serve:
	git-docs serve

.PHONY: build-docker
build-docker:
	docker build . -t git-docs

.PHONY: run-serve-docker
run-serve-docker:
	docker run -it --rm -v "$(shell pwd)":/usr/src/myapp -w /usr/src/myapp --user $(shell id -u):$(shell id -g) git-docs serve

.PHONY: run-serve-interactive
run-docker-interactive:
	docker run -it --rm -v "$(shell pwd)":/usr/src/myapp -w /usr/src/myapp --user $(shell id -u):$(shell id -g) --entrypoint sh git-docs
