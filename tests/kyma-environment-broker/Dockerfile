FROM golang:1.12.5-alpine3.9 as builder

ENV SRC_DIR=/go/src/github.com/kyma-incubator/compass/tests/kyma-environment-broker

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test

FROM alpine:3.10
RUN apk --no-cache add ca-certificates curl

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/kyma-environment-broker/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/kyma-environment-broker/test.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/kyma-environment-broker/licenses ./licenses

LABEL source=git@github.com:kyma-incubator/kyma-environment-broker.git

ENTRYPOINT ./entrypoint.sh
