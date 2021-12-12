FROM alpine:3.14.2

ARG MIGRATE_VER=4.14.1

WORKDIR /migrate

RUN apk --no-cache add bash postgresql-client curl jq
RUN wget https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VER}/migrate.linux-amd64.tar.gz -O - | tar -xz
RUN mv migrate.linux-amd64 /usr/local/bin/migrate

ENV CLUSTER_VERSION=1.19.8

RUN apk add --update ca-certificates \
 && apk add -t deps \
 && apk add --update curl \
 && export ARCH="$(uname -m)" && if [[ ${ARCH} == "x86_64" ]]; then export ARCH="amd64"; fi && curl -L https://dl.k8s.io/release/v${CLUSTER_VERSION}/bin/linux/${ARCH}/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl \
 && apk del --purge deps \
 && rm /var/cache/apk/*

COPY ./migrations/ ./migrations
COPY ./seeds/ ./seeds
COPY ./run.sh ./run.sh
COPY ./update-expected-schema-version.sh ./update-expected-schema-version.sh

ENTRYPOINT ["./run.sh"]
