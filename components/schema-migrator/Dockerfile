FROM alpine:3.16.2

ARG MIGRATE_VER=4.15.2
ARG TARGETARCH

WORKDIR /migrate

RUN apk --no-cache add bash postgresql-client curl jq
RUN wget https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VER}/migrate.linux-${TARGETARCH}.tar.gz -O - | tar -xz
RUN mv migrate /usr/local/bin/migrate

# kubectl is supported within one minor version (older or newer) of kube-apiserver
ENV CLUSTER_VERSION=1.21.12

RUN apk add --update ca-certificates \
 && apk add -t deps \
 && apk add --update curl \
 && curl -L https://dl.k8s.io/release/v${CLUSTER_VERSION}/bin/linux/${TARGETARCH}/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl \
 && apk del --purge deps \
 && rm /var/cache/apk/*

COPY ./migrations/ ./migrations
COPY ./seeds/ ./seeds
COPY ./run.sh ./run.sh
COPY ./update-expected-schema-version.sh ./update-expected-schema-version.sh

ENTRYPOINT ["./run.sh"]
