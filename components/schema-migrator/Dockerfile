FROM alpine:3.11.3

ARG MIGRATE_VER=4.5.0

WORKDIR /migrate

RUN apk --no-cache add bash postgresql-client curl
RUN wget https://github.com/golang-migrate/migrate/releases/download/v${MIGRATE_VER}/migrate.linux-amd64.tar.gz -O - | tar -xz
RUN mv migrate.linux-amd64 /usr/local/bin/migrate

COPY ./migrations/ ./migrations
COPY ./run.sh ./run.sh

ENTRYPOINT ["./run.sh"]
