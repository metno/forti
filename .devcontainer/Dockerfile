FROM mcr.microsoft.com/devcontainers/base:ubuntu-24.04


RUN apt-get update && apt-get install -y \
    gcc g++ pkg-config \
    git make \
    build-essential \
    cmake libgflags-dev libgoogle-glog-dev libgtest-dev libssl-dev

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

RUN apt-get update && apt-get install -y \
    libgdal-dev \
    libgeos-dev libproj-dev \
    libnetcdf-dev \
    shellcheck


COPY rawdataforecaster/build/package/golang_checksums.txt /tmp/golang_checksums.txt

# NOTE: When updating this, make sure to also set golang version in go.mod to the same version
ARG GOLANG_VERSION=1.26.1
ARG TARGETARCH

RUN \
    curl -L -O https://go.dev/dl/go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz && \
    GOLANG_SHA256=$(grep "go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz" /tmp/golang_checksums.txt | awk '{print $1}') && \
    sha256sum go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz | grep -q ${GOLANG_SHA256} && \
    tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz && \
    rm go${GOLANG_VERSION}.linux-${TARGETARCH}.tar.gz

RUN /usr/local/go/bin/go install golang.org/x/vuln/cmd/govulncheck@latest

RUN ln -s /usr/local/go/bin/go /usr/local/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt && \
    ln -s /root/go/bin/govulncheck /usr/local/bin/govulncheck
