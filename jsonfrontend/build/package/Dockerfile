FROM golang:1.26.4

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY internal/ /src/internal/
COPY jsonfrontend/internal /src/jsonfrontend/internal
COPY jsonfrontend/pkg /src/jsonfrontend/pkg

RUN go test ./...

COPY jsonfrontend/cmd /src/jsonfrontend/cmd

WORKDIR /src/jsonfrontend/cmd/jsonfrontend
RUN go build


FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

EXPOSE 8080

COPY jsonfrontend/cmd/jsonfrontend/jsonformat.json /
COPY --from=0 /src/jsonfrontend/cmd/jsonfrontend/jsonfrontend /usr/local/bin

USER nobody

ENTRYPOINT [ "/usr/local/bin/jsonfrontend" ]
