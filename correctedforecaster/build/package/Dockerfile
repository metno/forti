# FROM golang:1.15

# COPY go.mod /src/go.mod
# COPY go.sum /src/go.sum
# WORKDIR /src
# RUN go mod download

# COPY internal/ /src/internal/
# COPY pkg/ /src/pkg/
# COPY cmd/ /src/cmd

# RUN go test ./...

# WORKDIR /src/cmd/corrector
# RUN go build


FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y libgdal26 curl

RUN curl -L -o /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64 && \
    sha256sum /bin/grpc_health_probe | grep -q c78e988a4aad5e9e599c6a69e681ac68579c000b8f0571593325ccbc0c1638b7 && \
    chmod +x /bin/grpc_health_probe


COPY cmd/corrector/corrector /usr/local/bin
# COPY --from=0 /src/cmd/corrector/corrector /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/corrector" ]
