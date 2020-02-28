FROM golang:1.13.8-alpine3.11 as builder

ENV SRC_DIR=/go/src/github.com/kyma-incubator/compass/tests/connector-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/apitests

# After the next release of alpine is available (3.12) we should set it here instead of the edge tag
# We can't do it now cause alpine:3.11 contains a vulnerable version of curl
FROM alpine:edge

LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache add ca-certificates curl

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/apitests.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
