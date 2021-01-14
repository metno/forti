FROM ubuntu:focal

RUN apt-get update && apt-get -y install git curl

RUN \
    curl -L -O https://golang.org/dl/go1.15.5.linux-amd64.tar.gz && \
    sha256sum go1.15.5.linux-amd64.tar.gz | grep -q 9a58494e8da722c3aef248c9227b0e9c528c7318309827780f16220998180a0d && \
    tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz && \
    rm go1.15.5.linux-amd64.tar.gz

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
