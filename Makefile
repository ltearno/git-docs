.PHONY: build-image run-interactive

build:
	go build -o bin/mgit main.go

run:
	go run main.go

build-docker:
	docker build . -t magic-git

run-docker:
	docker run -it --rm -v "$(shell pwd)":/usr/src/myapp -w /usr/src/myapp --user $(shell id -u):$(shell id -g) magic-git
