# stage for caching modules
FROM golang:1.12 as build_base

# build envs
ENV GOOS linux
ENV GOARCH amd64

# service configs
ENV PKG github.com/lillilli/geth_contract

RUN mkdir -p /go/src/${PKG}
WORKDIR /go/src/${PKG}

COPY go.mod go.sum Makefile ./

RUN GO111MODULE=off go get github.com/ethereum/go-ethereum

# install bins for go-ethereum devtools
RUN curl -sL https://deb.nodesource.com/setup_8.x | bash && apt-get install -y nodejs
RUN curl -o /usr/bin/solc -fL https://github.com/ethereum/solidity/releases/download/v0.5.9/solc-static-linux \
  && chmod u+x /usr/bin/solc

RUN cd ${GOPATH}/src/github.com/ethereum/go-ethereum && make && make devtools
RUN GO111MODULE=on go mod vendor


# build main stage
FROM build_base as service_builder

COPY . .
RUN make gen_contract && make setup && make build


# result container
FROM alpine as service_runner

ENV SERVICE_NAME api
ENV PKG github.com/lillilli/geth_contract

WORKDIR /root/

COPY --from=service_builder /go/src/${PKG}/cmd/${SERVICE_NAME}/ .

RUN apk add --no-cache libc6-compat
RUN apk add --no-cache libcurl

EXPOSE 8080
ENTRYPOINT ["./api", "-config=local.yml"]