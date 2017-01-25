FROM golang:latest

ADD ./fsrestarter /fsrestarter

WORKDIR /

ENTRYPOINT ["/fsrestarter"]
