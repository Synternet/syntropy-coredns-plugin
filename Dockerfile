FROM golang:1.14 AS builder

WORKDIR /go/src/github.com/coredns/
RUN git clone https://github.com/coredns/coredns
WORKDIR /go/src/github.com/coredns/coredns
RUN mkdir -p plugin/syntropy
COPY ./src/* plugin/syntropy/
RUN sed -i 's/acl:acl/acl:acl\nsyntropy:syntropy/g' plugin.cfg # add Syntropy to plugin compile list
RUN make

FROM debian:stable-slim AS certs

RUN apt-get update && apt-get -uy upgrade
RUN apt-get install -y ca-certificates && update-ca-certificates 

FROM scratch
ENV LOCAL_CACHE_DURATION 300s

COPY --from=certs /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /go/src/github.com/coredns/coredns/coredns /coredns
ADD ./Corefile.docker /Corefile

EXPOSE 53 53/udp
ENTRYPOINT ["/coredns"]
