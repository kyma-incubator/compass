FROM golang:1.14.4-alpine3.12 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/control-plane/tests/e2e/provisioning

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -ldflags="-s -w" -c ./test

FROM alpine:3.12.0
LABEL source=git@github.com:kyma-project/control-plane.git

RUN apk --no-cache add ca-certificates curl

COPY --from=builder /go/src/github.com/kyma-project/control-plane/tests/e2e/provisioning/test.test .
COPY --from=builder /go/src/github.com/kyma-project/control-plane/tests/e2e/provisioning/licenses ./licenses

ENTRYPOINT test.test
