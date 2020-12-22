# FROM golang:1.15

# COPY go.mod /src/go.mod
# COPY go.sum /src/go.sum
# WORKDIR /src
# RUN go mod download

# COPY internal/ /src/internal/
# COPY pkg/ /src/pkg/
# COPY cmd/ /src/cmd

# RUN go test ./...

# WORKDIR /src/cmd/xmlfrontend
# RUN go build


FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

EXPOSE 8080

COPY cmd/xmlfrontend/xmlformat.json /
COPY cmd/xmlfrontend/xmlfrontend /usr/local/bin
# COPY --from=0 /src/cmd/xmlfrontend/xmlfrontend /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/xmlfrontend" ]
