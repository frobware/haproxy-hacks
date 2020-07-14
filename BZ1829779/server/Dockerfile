FROM docker.io/centos:7
USER root
COPY libwebsockets/build/bin/libwebsockets-test-server /
EXPOSE 7681
ENTRYPOINT ["/libwebsockets-test-server"]
