FROM golang:1.21

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY moxfrontend/internal /src/moxfrontend/internal
COPY moxfrontend/pkg /src/moxfrontend/pkg

RUN go test ./...

COPY moxfrontend/cmd /src/moxfrontend/cmd

WORKDIR /src/moxfrontend/cmd/moxfrontend
RUN go build


FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

EXPOSE 8080

COPY --from=0 /src/moxfrontend/cmd/moxfrontend/moxfrontend /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/moxfrontend" ]
