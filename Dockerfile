FROM golang:1.12-alpine

RUN mkdir /.cache && chmod ugo+rw /.cache