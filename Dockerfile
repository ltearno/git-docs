FROM golang:1.12-alpine

RUN mkdir /.cache && chmod ugo+rw /.cache

ADD main.go ./
ADD repository ./repository/
ADD tools ./tools/
ADD webserver ./webserver/

RUN go build -o mgit main.go

ENTRYPOINT [ "/go/mgit" ]