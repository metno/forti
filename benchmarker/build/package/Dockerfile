FROM golang:1.25

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY benchmarker/internal /src/benchmarker/internal

COPY benchmarker/cmd /src/benchmarker/cmd

WORKDIR /src/benchmarker/cmd/benchmarker
RUN go build


FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
COPY --from=0 /src/benchmarker/cmd/benchmarker/benchmarker /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/benchmarker" ]
