FROM ubuntu:24.04

ARG GOLANG_VERSION=1.26.1
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl ca-certificates

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

RUN apt-get update && apt-get install -y --no-install-recommends \
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


FROM ubuntu:24.04

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgdal34t64 ca-certificates

COPY --from=0 /src/correctedforecaster/cmd/correctedforecaster/correctedforecaster /usr/local/bin

USER nobody

ENTRYPOINT [ "/usr/local/bin/correctedforecaster" ]
