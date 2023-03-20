FROM ubuntu:jammy

ARG GOLANG_VERSION=1.20.2
ARG GOLANG_SHA256=4eaea32f59cde4dc635fbc42161031d13e1c780b87097f4b4234cfce671f1768
ARG ARCH=amd64

RUN apt-get update && apt-get -y install git curl

RUN \
    curl -L -O https://golang.org/dl/go${GOLANG_VERSION}.linux-${ARCH}.tar.gz && \
    sha256sum go${GOLANG_VERSION}.linux-${ARCH}.tar.gz | grep -q ${GOLANG_VERSION} && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-${ARCH}.tar.gz && \
    rm go${GOLANG_VERSION}.linux-${ARCH}.tar.gz

ENV GOPATH=/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    gcc g++ pkg-config \
    libgdal-dev

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY parameters/ /src/parameters/
COPY correctedforecaster/internal/ /src/correctedforecaster/internal/

RUN go test ./...

COPY correctedforecaster/cmd/ /src/correctedforecaster/cmd

WORKDIR /src/correctedforecaster/cmd/correctedforecaster
RUN go build


FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y libgdal26 curl

RUN curl -L -o /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64 && \
    sha256sum /bin/grpc_health_probe | grep -q c78e988a4aad5e9e599c6a69e681ac68579c000b8f0571593325ccbc0c1638b7 && \
    chmod +x /bin/grpc_health_probe


COPY --from=0 /src/correctedforecaster/cmd/correctedforecaster/correctedforecaster /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/correctedforecaster" ]
