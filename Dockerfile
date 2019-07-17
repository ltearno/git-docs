FROM golang:1.12-alpine

RUN mkdir /.cache && chmod ugo+rw /.cache

ADD main.go Makefile ./

RUN go build -o mgit main.go