FROM ubuntu:18.04
LABEL maintainer="jishankai"

#prerequisite
RUN apt-get update && apt-get install -y \
    vim-tiny \
    curl \
    git \
    build-essential \
    make \
    g++ \
    swig \
    libbz2-dev \
    libgflags-dev \
    liblz4-dev \
    libsnappy-dev \
    libzstd-dev \
    zlib1g-dev \
 && rm -rf /var/lib/apt/lists/*

# install golang
WORKDIR /code
RUN curl -sSL https://storage.googleapis.com/golang/go1.12.10.linux-amd64.tar.gz \
		| tar -v -C /usr/local -xz
ENV GOROOT /usr/local/go
ENV GOPATH $HOME/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH
ENV GOPROXY https://goproxy.io
ENV GO111MODULE on

# install QuarkChain-Mining-Pool
WORKDIR $GOPATH/src/github.com/QuarkChain-Mining-Pool
RUN git clone https://github.com/QuarkChain/QuarkChain-Mining-Pool.git
WORKDIR $GOPATH/src/github.com/QuarkChain/QuarkChain-Mining-Pool
RUN git checkout api-potimization
RUN go build
