FROM golang:1.12-alpine

RUN apk add make git

RUN mkdir /.cache && chmod ugo+rw /.cache

ADD Makefile main.go ./
ADD repository ./repository/
ADD tools ./tools/
ADD webserver ./webserver/
ADD assets ./assets/

RUN make build-prepare
RUN make build-embed-assets
RUN make install

ENTRYPOINT [ "/go/mgit" ]