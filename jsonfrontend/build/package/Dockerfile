# FROM golang:1.15

# COPY go.mod /src/go.mod
# COPY go.sum /src/go.sum
# WORKDIR /src
# RUN go mod download

# COPY internal/ /src/internal/
# COPY pkg/ /src/pkg/
# COPY cmd/ /src/cmd

# RUN go test ./...

# WORKDIR /src/cmd/jsonfrontend
# RUN go build


FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

EXPOSE 8080

COPY cmd/jsonfrontend/jsonformat.json /
COPY cmd/jsonfrontend/jsonfrontend /usr/local/bin
# COPY --from=0 /src/cmd/jsonfrontend/jsonfrontend /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/jsonfrontend" ]
