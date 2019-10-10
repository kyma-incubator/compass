FROM golang:1.12.5-alpine3.9 as builder

ENV SRC_DIR=/go/src/github.com/kyma-incubator/compass/tests/connector-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/apitests

FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/apitests.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh