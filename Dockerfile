FROM golang:latest

ADD ./fsrestarter /fsrestarter

ENTRYPOINT ["/fsrestarter"]
