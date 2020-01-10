FROM golang:1.13.5-alpine3.10 as builder


ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests/director

WORKDIR ${BASE_TEST_DIR}
COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./api-tests && \
    CGO_ENABLED=0 go test -c ./gateway-integration-tests

FROM alpine:3.10

RUN apk add --no-cache curl

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director/wait-for-director.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director/api-tests.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director/gateway-integration-tests.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director/licenses ./licenses

