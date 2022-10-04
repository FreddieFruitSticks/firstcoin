# syntax=docker/dockerfile:1

FROM golang:1.17.7-alpine3.15

ARG HOST_NAME

ADD . /go/src/firstcoin

ENV GOPATH=/go
ENV HOST_NAME=$HOST_NAME
ENV PATH="/go/bin:${PATH}" 

WORKDIR /go/src/firstcoin

RUN go install ./... 

# CMD ["/go/bin/firstcoin", "$NODE_PORT"]
# CMD ["echo", "$NODE_PORT"]
# CMD ["sh", "-c", "/go/bin/firstcoin $NODE_PORT $PEER_PORT"]
