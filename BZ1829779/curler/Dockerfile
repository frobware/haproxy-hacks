FROM ubi8 as builder
RUN yum install -y gcc gcc-c++ make libcurl-devel cmake pkgconfig
WORKDIR /work
RUN mkdir /target
COPY . .
RUN (mkdir build; cmake ..; make clean; make all)
CMD cp build/curler /target
