FROM centos as builder
RUN yum install -y gcc gcc-c++ make libcurl-devel cmake pkgconfig
WORKDIR /work
RUN mkdir /target
COPY . .
RUN (mkdir build; cmake ..; make)
CMD cp build/curler /target
