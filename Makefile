.PHONY: build-image run-interactive

build-prepare:
	go get -u github.com/jteeuwen/go-bindata/...
	go get github.com/julienschmidt/httprouter

build-embed-assets:
	go-bindata -o assetsgen/assets.go -pkg assetsgen assets/...

build: build-embed-assets
	go install git-docs

run-serve:
	git-docs serve

build-docker:
	docker build . -t git-docs

run-serve-docker:
	docker run -it --rm -v "$(shell pwd)":/usr/src/myapp -w /usr/src/myapp --user $(shell id -u):$(shell id -g) git-docs serve

run-docker-interactive:
	docker run -it --rm -v "$(shell pwd)":/usr/src/myapp -w /usr/src/myapp --user $(shell id -u):$(shell id -g) --entrypoint sh git-docs
