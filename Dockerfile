FROM golang:latest

ADD ./fsrestarter /bin/fsrestarter

WORKDIR /

ENTRYPOINT ["/bin/fsrestarter"]
