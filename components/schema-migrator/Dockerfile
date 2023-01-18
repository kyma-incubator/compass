FROM debian:11.6-slim

ARG MIGRATE_VER=4.15.2
ARG TARGETARCH

WORKDIR /migrate

RUN ln -s /usr/bin/dpkg-split /usr/sbin/dpkg-split
RUN ln -s /usr/bin/dpkg-deb /usr/sbin/dpkg-deb
RUN ln -s /bin/rm /usr/sbin/rm
RUN ln -s /bin/tar /usr/sbin/tar

RUN export DEBIAN_FRONTEND=noninteractive apt-get update && \
    apt-get install -y postgresql-client jq curl && \
    rm -rf /var/lib/apt/lists/*

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VER}/migrate.linux-${TARGETARCH}.tar.gz | tar -xz
RUN mv migrate /usr/local/bin/migrate

# kubectl is supported within one minor version (older or newer) of kube-apiserver
ENV CLUSTER_VERSION=1.21.12

RUN curl -L https://dl.k8s.io/release/v${CLUSTER_VERSION}/bin/linux/${TARGETARCH}/kubectl -o /usr/local/bin/kubectl  && chmod +x /usr/local/bin/kubectl 

COPY ./migrations/ ./migrations
COPY ./seeds/ ./seeds
COPY ./run.sh ./run.sh
COPY ./update-expected-schema-version.sh ./update-expected-schema-version.sh

ENTRYPOINT ["./run.sh"]
