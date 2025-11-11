FROM ubuntu:24.04
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
    libnetcdf-dev

WORKDIR /src

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY fortiup/internal /src/fortiup/internal 
COPY fortiup/pkg /src/fortiup/pkg

RUN go test ./...

COPY fortiup/cmd /src/fortiup/cmd

WORKDIR /src/fortiup/cmd/fortiup
RUN go build

FROM ubuntu:24.04

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y libnetcdf19t64

COPY --from=0 /src/fortiup/cmd/fortiup/fortiup /usr/local/bin/

# Make sure that it runs
RUN ldd /usr/local/bin/fortiup
