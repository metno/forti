FROM ubuntu:focal

ARG GOLANG_VERSION=1.17.1
ARG GOLANG_SHA256=dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef

RUN apt-get update && apt-get -y install git curl

RUN \
    curl -L -O https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    sha256sum go${GOLANG_VERSION}.linux-amd64.tar.gz | grep -q ${GOLANG_VERSION} && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    rm go${GOLANG_VERSION}.linux-amd64.tar.gz

ENV GOPATH=/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    gcc g++ pkg-config \
    cmake libgflags-dev libgoogle-glog-dev libgtest-dev libssl-dev \
    libgeos-dev libproj-dev


WORKDIR /src
RUN git clone https://github.com/google/s2geometry.git && \
    cd s2geometry && \
    git checkout 2c02e21040e0b82aa5719e96033d02b8ce7c0eff && \
    mkdir build && \
    cd build/ && \
    cmake .. && \
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


FROM ubuntu:focal

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    libgflags2.2 libgoogle-glog0v5 libssl1.1 \
    libgeos-c1v5 libproj15 \
    curl

RUN curl -L -o /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64 && \
    sha256sum /bin/grpc_health_probe | grep -q c78e988a4aad5e9e599c6a69e681ac68579c000b8f0571593325ccbc0c1638b7 && \
    chmod +x /bin/grpc_health_probe

EXPOSE 8080

COPY --from=0 /usr/local/lib/libs2.so /usr/local/lib/
RUN ldconfig


COPY --from=0 /src/rawdataforecaster/cmd/rawdataforecaster/rawdataforecaster /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/rawdataforecaster" ]
