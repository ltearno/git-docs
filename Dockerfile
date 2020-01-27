FROM golang:1.12-alpine

RUN apk add make git

RUN mkdir /.cache && chmod ugo+rw /.cache

RUN mkdir -p /test-dir

ADD Makefile ./
ADD src ./src/
ADD assets ./assets/

RUN make build-prepare
RUN make build-embed-assets
RUN make install

ENTRYPOINT [ "/go/bin/git-docs" ]