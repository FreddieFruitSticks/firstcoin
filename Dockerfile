# syntax=docker/dockerfile:1

FROM golang:1.17.7-alpine3.15

ADD . /go/src/blockchain
COPY ./wait-for /go/src/blockchain

ENV GOPATH=/go/src/

WORKDIR /go/src/blockchain

RUN go install ./...

ENV PATH="/go/src/bin:${PATH}" 


# CMD "./wait-for-it.sh" $PEER_PORT -- "blockchain" $NODE_PORT $PEER_PORT
