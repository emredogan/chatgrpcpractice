FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /chat
RUN mkdir -p /chat/proto

WORKDIR /chat

COPY ./proto/service.pb.go /chat/proto
COPY ./proto/service_grpc.pb.go /chat/proto
COPY ./server/main.go /chat

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o chat

CMD ./chat