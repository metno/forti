FROM ubuntu:22.04

ARG GOLANG_VERSION=1.20.5
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
    cmake libgflags-dev libgoogle-glog-dev libgtest-dev libssl-dev \
    libgeos-dev libproj-dev


WORKDIR /src

RUN git clone https://github.com/abseil/abseil-cpp.git && \
    cd abseil-cpp && \
    git checkout c2435f8342c2d0ed8101cb43adfd605fdc52dca2 && \
    mkdir build && \
    cd build/ && \
    cmake -DCMAKE_CXX_STANDARD=17 -DCMAKE_POSITION_INDEPENDENT_CODE=ON .. && \
    make && \
    make install

RUN git clone https://github.com/google/s2geometry.git && \
    cd s2geometry && \
    git checkout 7773d518b1f29caa1c2045eb66ec519e025be108 && \
    mkdir build && \
    cd build/ && \
    cmake  -DCMAKE_CXX_STANDARD=17 .. && \
    make && \
    make install

RUN ldconfig


COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
WORKDIR /src
RUN go mod download

COPY internalprotocol /src/internalprotocol
COPY upload/pkg /src/upload/pkg
COPY rawdataforecaster/internal/ /src/rawdataforecaster/internal
COPY rawdataforecaster/test/ /src/rawdataforecaster/test

RUN go test ./...

COPY rawdataforecaster/cmd /src/rawdataforecaster/cmd

WORKDIR /src/rawdataforecaster/cmd/rawdataforecaster
RUN go build


FROM ubuntu:22.04

ARG TARGETARCH

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    libgflags2.2 libgoogle-glog0v5 libssl3 \
    libgeos-c1v5 libproj22 \
    curl

COPY rawdataforecaster/build/package/grpc_health_probe_checksums.txt /tmp/grpc_health_probe_checksums.txt
RUN curl -L -o /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.19/grpc_health_probe-linux-${TARGETARCH} && \
    sha256sum /bin/grpc_health_probe | grep -q $(grep "grpc_health_probe-linux-${TARGETARCH}" /tmp/grpc_health_probe_checksums.txt | awk '{print $1}') && \
    chmod +x /bin/grpc_health_probe

EXPOSE 8080

COPY --from=0 /usr/local/lib/libs2.so /usr/local/lib/
RUN ldconfig


COPY --from=0 /src/rawdataforecaster/cmd/rawdataforecaster/rawdataforecaster /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/rawdataforecaster" ]
