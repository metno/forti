FROM ubuntu:24.04

ARG GOLANG_VERSION=1.26.4
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
    build-essential \
    gcc g++ pkg-config \
    cmake libgflags-dev libgoogle-glog-dev libgtest-dev libssl-dev \
    libgeos-dev libproj-dev

WORKDIR /src

RUN git clone https://github.com/abseil/abseil-cpp.git
WORKDIR /src/abseil-cpp
RUN git checkout c2435f8342c2d0ed8101cb43adfd605fdc52dca2
WORKDIR /src/abseil-cpp/build
RUN cmake -DCMAKE_CXX_STANDARD=17 -DCMAKE_POSITION_INDEPENDENT_CODE=ON .. && \
    make && \
    make install

WORKDIR /src
RUN git clone https://github.com/google/s2geometry.git
WORKDIR /src/s2geometry
RUN git checkout 7773d518b1f29caa1c2045eb66ec519e025be108
WORKDIR /src/s2geometry/build
RUN cmake  -DCMAKE_CXX_STANDARD=17 .. && \
    make && \
    make install

RUN ldconfig

COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY rawdataforecaster/internal/ /src/rawdataforecaster/internal
COPY rawdataforecaster/test/ /src/rawdataforecaster/test

RUN go test ./...

COPY rawdataforecaster/cmd /src/rawdataforecaster/cmd

WORKDIR /src/rawdataforecaster/cmd/rawdataforecaster
RUN go build


FROM ubuntu:24.04

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    libgflags2.2 libgoogle-glog0v6t64 libssl3 \
    libgeos-c1v5 libproj25 \
    ca-certificates

EXPOSE 8080

COPY --from=0 /usr/local/lib/libs2.so /usr/local/lib/
RUN ldconfig

COPY --from=0 /src/rawdataforecaster/cmd/rawdataforecaster/rawdataforecaster /usr/local/bin

USER nobody

ENTRYPOINT [ "/usr/local/bin/rawdataforecaster" ]
