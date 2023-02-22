FROM golang:1.20.0-alpine3.17 as builder

WORKDIR /
RUN apk --update --no-cache add bash git

# Currently the commit hash of the version, as there hasn't been a release since April 2022
ARG VERSION="dc26c41ac9dac4a1023c87723ffda644e41af1fe"
ARG DATABASE="postgres"
ARG SOURCE="file"

RUN git clone https://github.com/golang-migrate/migrate
WORKDIR /migrate
RUN git checkout $VERSION


RUN go build -a -o migrate -ldflags="-X main.Version=$VERSION" -tags "$DATABASE $SOURCE" ./cmd/migrate

FROM alpine:3.17.1

WORKDIR /migrate

RUN apk --update --no-cache add bash postgresql-client curl jq

COPY --from=builder /migrate /migrate
RUN chmod +x ./migrate

RUN mv migrate /usr/local/bin/migrate

# kubectl is supported within one minor version (older or newer) of kube-apiserver
ENV CLUSTER_VERSION=1.26.1

RUN curl -L https://dl.k8s.io/release/v${CLUSTER_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl  && chmod +x /usr/local/bin/kubectl

COPY ./migrations/ ./migrations
COPY ./seeds/ ./seeds
COPY ./run.sh ./run.sh
COPY ./update-expected-schema-version.sh ./update-expected-schema-version.sh

ENTRYPOINT ["./run.sh"]
