FROM busybox

ADD ./fsrestarter fsrestarter
ADD ./target /target
RUN rm -rf /target/*

ENTRYPOINT ["/fsrestarter"]
