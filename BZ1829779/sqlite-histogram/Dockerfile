FROM centos/devtoolset-7-toolchain-centos7
USER root
WORKDIR /workdir
ADD sqlite-amalgamation-3320300 /workdir
RUN make
ADD installer /installer
CMD /installer /target
