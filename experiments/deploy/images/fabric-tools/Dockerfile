ARG UBUNTU_VER=24.04
FROM docker.io/library/ubuntu:${UBUNTU_VER} AS builder

ARG TARGETARCH=amd64
ARG TARGETOS=linux
ARG FABRIC_VER=3.1.4
ARG GO_VER=1.26.0
ENV DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC

RUN apt update && apt install -y \
    git \
    gcc \
    curl \
    make \
    vim

RUN curl -sL https://go.dev/dl/go${GO_VER}.${TARGETOS}-${TARGETARCH}.tar.gz | tar zxf - -C /usr/local
ENV PATH="/usr/local/go/bin:$PATH"

RUN git clone https://github.com/hyperledger/fabric.git

RUN cd fabric && \
    git checkout v${FABRIC_VER} && \
    make native FABRIC_VER=${FABRIC_VER}


###############################################################################
# Runtime image
###############################################################################

ARG UBUNTU_VER=24.04
FROM docker.io/library/ubuntu:${UBUNTU_VER}

ARG TARGETARCH=amd64
ARG TARGETOS=linux
ARG FABRIC_VER=3.1.4
ARG GO_VER=1.26.0

RUN apt update && apt install -y \
    bash \
    curl \
    jq \
    libssl-dev \
    tzdata

RUN curl -sL https://go.dev/dl/go${GO_VER}.${TARGETOS}-${TARGETARCH}.tar.gz | tar zxvf - -C /usr/local
ENV PATH="/usr/local/go/bin:$PATH"

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN echo 'hosts: files dns' > /etc/nsswitch.conf

ENV     FABRIC_CFG_PATH=/etc/hyperledger/fabric
ENV     FABRIC_VER=${FABRIC_VER}

COPY    --from=builder  fabric/sampleconfig    ${FABRIC_CFG_PATH}
COPY    --from=builder  fabric/build/bin       /usr/local/bin

VOLUME  /etc/hyperledger/fabric
VOLUME  /var/hyperledger

