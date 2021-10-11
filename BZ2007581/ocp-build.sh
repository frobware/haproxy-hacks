#!/usr/bin/env bash

set -eux

make CPU="generic" \
     TARGET="linux-glibc" \
     USE_OPENSSL=1 \
     USE_PCRE=1 \
     USE_ZLIB=1 \
     USE_CRYPT_H=1 \
     USE_LINUX_TPROXY=1 \
     USE_GETADDRINFO=1 \
     V=1 \
     EXTRA_OBJS="contrib/interposer/accept.c contrib/interposer/malloc.c" \
     DEBUG_CFLAGS="-Og -ggdb3" "$@"

# The last three lines are additions; OCP does not build with debug or
# the contrib extras.
