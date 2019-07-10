# Build stage
FROM golang:1.12-stretch as build-env

RUN mkdir -p /go/src/github.com/gluster/gluster-prometheus/

WORKDIR /go/src/github.com/gluster/gluster-prometheus/

RUN apt-get -q update && apt-get install -y bash curl make git

COPY . .

RUN scripts/install-reqs.sh
RUN PREFIX=/app make
RUN PREFIX=/app make install

# Create small image for running
FROM debian:stretch

ARG GLUSTER_VERSION=6

# Install gluster cli for gluster-exporter
RUN apt-get -q update && apt-get install -y gnupg curl apt-transport-https && \
        DEBID=$(grep 'VERSION_ID=' /etc/os-release | cut -d '=' -f 2 | tr -d '"') && \
        DEBVER=$(grep 'VERSION=' /etc/os-release | grep -Eo '[a-z]+') && \
        DEBARCH=$(dpkg --print-architecture) && \
        curl -sSL http://download.gluster.org/pub/gluster/glusterfs/${GLUSTER_VERSION}/rsa.pub | apt-key add - && \
        echo deb https://download.gluster.org/pub/gluster/glusterfs/${GLUSTER_VERSION}/LATEST/Debian/${DEBID}/${DEBARCH}/apt ${DEBVER} main > /etc/apt/sources.list.d/gluster.list && \
        apt-get -q update && apt-get install -y glusterfs-server && apt-get clean all

WORKDIR /app

COPY --from=build-env /app /app/

ENTRYPOINT ["/app/sbin/gluster-exporter"]
