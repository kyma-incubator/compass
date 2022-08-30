FROM golang:1.18.5-alpine3.16 as connectivity-adapter-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./connectivity-adapter/tests -o connectivity-adapter.test

FROM golang:1.18.5-alpine3.16 as connector-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./connector/tests -o connector.test

FROM golang:1.18.5-alpine3.16 as director-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./director/tests -o director.test

FROM golang:1.18.5-alpine3.16 as ns-adapter-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./ns-adapter/tests -o ns-adapter.test

FROM golang:1.18.5-alpine3.16 as external-services-mock-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./external-services-mock/tests -o external-services-mock.test

FROM golang:1.18.5-alpine3.16 as gateway-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./gateway/tests -o gateway.test

FROM golang:1.18.5-alpine3.16 as ord-aggregator-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./ord-aggregator/tests -o ord-aggregator.test

FROM golang:1.18.5-alpine3.16 as ord-service-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./ord-service/tests -o ord-service.test

FROM golang:1.18.5-alpine3.16 as system-broker-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./system-broker/tests -o system-broker.test

FROM golang:1.18.5-alpine3.16 as tenant-fetcher-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./tenant-fetcher/tests -o tenant-fetcher.test

FROM golang:1.18.5-alpine3.16 as system-fetcher-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./system-fetcher/tests -o system-fetcher.test

FROM golang:1.18.5-alpine3.16 as istio-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./istio/tests -o istio.test

FROM golang:1.18.5-alpine3.16 as pairing-adapter-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./pairing-adapter/tests -o pairing-adapter.test

FROM golang:1.18.5-alpine3.16 as director-bench-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./director/bench -o director.bench

FROM golang:1.18.5-alpine3.16 as ord-service-bench-builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./ord-service/bench -o ord-service.bench

FROM alpine:3.16.2

RUN apk add --no-cache curl

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=connectivity-adapter-builder /go/src/github.com/kyma-incubator/compass/tests/connectivity-adapter.test .
COPY --from=connector-builder /go/src/github.com/kyma-incubator/compass/tests/connector.test .
COPY --from=director-builder /go/src/github.com/kyma-incubator/compass/tests/director.test .
COPY --from=ns-adapter-builder /go/src/github.com/kyma-incubator/compass/tests/ns-adapter.test .
COPY --from=external-services-mock-builder /go/src/github.com/kyma-incubator/compass/tests/external-services-mock.test .
COPY --from=gateway-builder /go/src/github.com/kyma-incubator/compass/tests/gateway.test .
COPY --from=ord-aggregator-builder /go/src/github.com/kyma-incubator/compass/tests/ord-aggregator.test .
COPY --from=ord-service-builder /go/src/github.com/kyma-incubator/compass/tests/ord-service.test .
COPY --from=system-broker-builder /go/src/github.com/kyma-incubator/compass/tests/system-broker.test .
COPY --from=tenant-fetcher-builder /go/src/github.com/kyma-incubator/compass/tests/tenant-fetcher.test .
COPY --from=system-fetcher-builder /go/src/github.com/kyma-incubator/compass/tests/system-fetcher.test .
COPY --from=istio-builder /go/src/github.com/kyma-incubator/compass/tests/istio.test .
COPY --from=pairing-adapter-builder /go/src/github.com/kyma-incubator/compass/tests/pairing-adapter.test .

COPY --from=director-bench-builder /go/src/github.com/kyma-incubator/compass/tests/director.bench .
COPY --from=ord-service-bench-builder /go/src/github.com/kyma-incubator/compass/tests/ord-service.bench .
#COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/licenses ./licenses
