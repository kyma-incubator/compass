FROM alpine:3.13.5

ARG MIGRATE_VER=4.5.0

WORKDIR /migrate

RUN apk --no-cache add bash postgresql-client curl
RUN wget https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VER}/migrate.linux-amd64.tar.gz -O - | tar -xz
RUN mv migrate.linux-amd64 /usr/local/bin/migrate

ENV CLUSTER_VERSION=1.19.8

RUN apk add --update ca-certificates \
 && apk add -t deps \
 && apk add --update curl \
 && export ARCH="$(uname -m)" && if [[ ${ARCH} == "x86_64" ]]; then export ARCH="amd64"; fi && curl -L https://storage.googleapis.com/kubernetes-release/release/${CLUSTER_VERSION}/bin/linux/${ARCH}/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl \
 && apk del --purge deps \
 && rm /var/cache/apk/*

COPY ./migrations/ ./migrations
COPY ./seeds/ ./seeds
COPY ./run.sh ./run.sh

ENTRYPOINT ["./run.sh"]
