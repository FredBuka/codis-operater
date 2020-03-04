# Alpha mirroring can also be used, but parsing cluster service names is a bit problematic
FROM centos:7 as builder

ENV GOLANG_VERSION 1.12
ENV GOLANG_DOWNLOAD_URL https://dl.google.com/go/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOPATH /gopath
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN yum update -y                                                    \
 && yum install -y wget                                              \
                  tar                                                \
                  g++                                                \
                  gcc                                                \
                  libc6-dev                                          \
                  make                                               \
                  git                                                \
                  autoconf                                           \
 && curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz               \
 && tar -C /usr/local -xzf golang.tar.gz                             \
 && rm golang.tar.gz                                                 \
 && mkdir -p "$GOPATH/src/github.com/CodisLabs" "$GOPATH/bin"        \
 && chmod -R 777 "$GOPATH"                                           \
 && cd $GOPATH/src/github.com/CodisLabs                              \
 && git clone https://github.com/CodisLabs/codis.git                 \
 && cd $GOPATH/src/github.com/CodisLabs/codis                        \
 && make build-all

# Build codis image
FROM centos:7

# Setup tools
RUN yum update -y && yum install -y bind-utils

# Copy from builder
COPY --from=builder /gopath/src/github.com/CodisLabs/codis/config /data/service/codis/config-default
COPY --from=builder /gopath/src/github.com/CodisLabs/codis/bin /data/service/codis/bin

ENV PATH ${PATH}:/data/service/codis/bin

WORKDIR /data/service/codis
