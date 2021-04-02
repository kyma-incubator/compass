FROM golang:1.15.7-alpine3.13 as builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests
WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_TEST_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./connectivity-adapter/tests -o connectivity-adapter.test && \
    CGO_ENABLED=0 go test -c ./connector/tests -o connector.test && \
    CGO_ENABLED=0 go test -c ./director/tests -o director.test && \
    CGO_ENABLED=0 go test -c ./external-services-mock/tests -o external-services-mock.test && \
    CGO_ENABLED=0 go test -c ./gateway/tests -o gateway.test && \
    CGO_ENABLED=0 go test -c ./operations-controller/tests -o operations-controller.test && \
    CGO_ENABLED=0 go test -c ./ord-aggregator/tests -o ord-aggregator.test && \
    CGO_ENABLED=0 go test -c ./ord-service/tests -o ord-service.test && \
    CGO_ENABLED=0 go test -c ./system-broker/tests -o system-broker.test && \
    CGO_ENABLED=0 go test -c ./tenant-fetcher/tests -o tenant-fetcher.test && \
    CGO_ENABLED=0 go test -c ./istio/tests -o istio.test

FROM alpine:3.13.2

RUN apk add --no-cache curl

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connectivity-adapter.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/external-services-mock.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/gateway.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/operations-controller.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/ord-aggregator.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/ord-service.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/system-broker.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/tenant-fetcher.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/istio.test .
#COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/licenses ./licenses
