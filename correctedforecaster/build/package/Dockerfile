FROM ubuntu:22.04

ARG GOLANG_VERSION=1.25.4
ARG TARGETARCH

RUN apt-get update && apt-get -y install git curl

COPY rawdataforecaster/build/package/golang_checksums.txt /tmp/golang_checksums.txt

RUN \
    curl -L -O https://go.dev/dl/go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz && \
    GOLANG_SHA256=$(grep "go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz" /tmp/golang_checksums.txt | awk '{print $1}') && \
    sha256sum go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz | grep -q ${GOLANG_SHA256} && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz && \
    rm go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz

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


FROM ubuntu:22.04

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y libgdal30 curl

COPY correctedforecaster/build/package/grpc_health_probe_checksums.txt /tmp/grpc_health_probe_checksums.txt
RUN curl -L -o /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.19/grpc_health_probe-linux-${TARGETARCH} && \
    sha256sum /bin/grpc_health_probe | grep -q $(grep "grpc_health_probe-linux-${TARGETARCH}" /tmp/grpc_health_probe_checksums.txt | awk '{print $1}') && \
    chmod +x /bin/grpc_health_probe


COPY --from=0 /src/correctedforecaster/cmd/correctedforecaster/correctedforecaster /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/correctedforecaster" ]
