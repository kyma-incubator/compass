FROM golang:1.14.4-alpine3.12 as builder

ENV SRC_DIR=/go/src/github.com/kyma-project/control-plane/tests/provisioner-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/provisioner



FROM alpine:3.12.0
LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache add ca-certificates curl

COPY --from=builder /go/src/github.com/kyma-project/control-plane/tests/provisioner-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-project/control-plane/tests/provisioner-tests/provisioner.test .
COPY --from=builder /go/src/github.com/kyma-project/control-plane/tests/provisioner-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
