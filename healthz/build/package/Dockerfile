FROM golang:1.26.4

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY jsonfrontend/pkg/ /src/jsonfrontend/pkg/
COPY xmlfrontend/pkg/ /src/xmlfrontend/pkg/
COPY internalprotocol/ /src/internalprotocol/

COPY healthz /src/healthz

RUN go test /src/healthz/...

WORKDIR /src/healthz/cmd/healthz

RUN go build

FROM ubuntu:24.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates

RUN useradd -ms /bin/bash forti
USER forti

COPY --from=0 /src/healthz/cmd/healthz/healthz /usr/local/bin/

EXPOSE 8080

USER nobody

ENTRYPOINT [ "/usr/local/bin/healthz", "-run-server"]
