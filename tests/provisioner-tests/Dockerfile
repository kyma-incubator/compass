FROM golang:1.12.9-alpine3.10 as builder

ENV SRC_DIR=/go/src/github.com/kyma-incubator/compass/tests/provisioner-tests

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/provisioner

FROM alpine:3.10

LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/provisioner-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/provisioner-tests/provisioner.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/provisioner-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh