FROM golang:1.13.5-alpine3.10 as builder


ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests/director-tests

WORKDIR ${BASE_TEST_DIR}
COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./api-tests && \
    CGO_ENABLED=0 go test -c ./auth-related-tests

FROM alpine:3.10

RUN apk add --no-cache curl

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director-tests/wait-for-director.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director-tests/api-tests.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director-tests/auth-related-tests.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/director-tests/licenses ./licenses

