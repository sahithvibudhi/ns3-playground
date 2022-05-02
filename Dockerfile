FROM ubuntu:16.04

MAINTAINER Sahith Vibudhi <v.sahithkumar@gmail.com>

LABEL Description="Docker image for NS-3 Network Simulator"

RUN apt-get update

# Needed packages
RUN apt-get install -y \
  git \
  mercurial \
  gcc \
  g++ \
  vim \
  python \
  python-dev \
  python-setuptools \
  qt5-default \
  python-pygraphviz \
  python-kiwi \
  python-pygoocanvas \
  libgoocanvas-dev \
  ipython \
  autoconf \
  cvs \
  bzr \
  unrar \
  gdb \
  valgrind \
  uncrustify \
  flex \
  bison \
  libfl-dev \
  tcpdump \
  gsl-bin \
  libgsl2 \
  libgsl-dev \
  sqlite \
  sqlite3 \
  libsqlite3-dev \
  libxml2 \
  libxml2-dev \
  cmake \
  libc6-dev \
  libc6-dev-i386 \
  libclang-dev \
  llvm-dev \
  automake \
  libgtk2.0-0 \
  libgtk2.0-dev \
  vtun \
  lxc \
  libboost-signals-dev \
  libboost-filesystem-dev

# install curl 
RUN apt-get install curl
# get install script and pass it to execute: 
RUN curl -sL https://deb.nodesource.com/setup_4.x | bash
# and install node 
RUN apt-get install nodejs
# confirm that it was successful 
RUN node -v
# npm installs automatically 
RUN npm -v

# NS-3
# Create working directory
RUN mkdir -p /usr/ns3
WORKDIR /usr

RUN wget http://www.nsnam.org/release/ns-allinone-3.30.1.tar.bz2
RUN tar -xf ns-allinone-3.30.1.tar.bz2

# Configure and compile NS-3
RUN cd ns-allinone-3.30.1 && ./build.py --enable-examples --enable-tests

RUN ln -s /usr/ns-allinone-3.30.1/ns-3.30.1/ /usr/ns3/

# Cleanup
RUN apt-get clean && \
  rm -rf /var/lib/apt