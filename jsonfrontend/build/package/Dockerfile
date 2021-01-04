FROM golang:1.15

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY weathersymbol/ /src/weathersymbol/
COPY jsonfrontend/internal /src/jsonfrontend/internal
COPY jsonfrontend/pkg /src/jsonfrontend/pkg

RUN go test ./...

COPY jsonfrontend/cmd /src/jsonfrontend/cmd

WORKDIR /src/jsonfrontend/cmd/jsonfrontend
RUN go build


FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

EXPOSE 8080

COPY jsonfrontend/cmd/jsonfrontend/jsonformat.json /
COPY --from=0 /src/jsonfrontend/cmd/jsonfrontend/jsonfrontend /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/jsonfrontend" ]
