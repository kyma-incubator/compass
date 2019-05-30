FROM golang:1.12.5-alpine3.9 as builder


ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests/compass-integration

WORKDIR ${BASE_TEST_DIR}
COPY . ${BASE_TEST_DIR}

RUN go test -c .

FROM alpine:3.9

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/compass-integration/compass-integration.test .

CMD ["./compass-integration.test", "-test.v"]