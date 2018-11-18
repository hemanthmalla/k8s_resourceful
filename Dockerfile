FROM golang:1.11.2-alpine3.8

WORKDIR /go/src/app

COPY bin/server /go/src/app/

EXPOSE 50051

CMD ./server




