build-image:
	docker build . -t aaa-installer

run-interactive:
	docker run -it --rm -v "$(shell pwd)":/usr/src/myapp -w /usr/src/myapp --user $(shell id -u):$(shell id -g) aaa-installer

