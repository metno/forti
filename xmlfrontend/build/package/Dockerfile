FROM golang:1.15

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY weathersymbol/ /src/weathersymbol/
COPY xmlfrontend/internal /src/xmlfrontend/internal
COPY xmlfrontend/pkg /src/xmlfrontend/pkg

RUN go test ./...

COPY xmlfrontend/cmd /src/xmlfrontend/cmd

WORKDIR /src/xmlfrontend/cmd/xmlfrontend
RUN go build


FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

EXPOSE 8080

COPY xmlfrontend/cmd/xmlfrontend/xmlformat.json /
COPY --from=0 /src/xmlfrontend/cmd/xmlfrontend/xmlfrontend /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/xmlfrontend" ]
