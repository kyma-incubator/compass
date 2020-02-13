FROM golang:1.12.5-alpine3.9 as builder

ENV SRC_DIR=/go/src/github.com/kyma-incubator/compass/tests/e2e/provisioning

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -ldflags="-s -w" -c ./test

FROM scratch

COPY --from=builder $SRC_DIR/test.test .
COPY --from=builder $SRC_DIR/licenses ./licenses

LABEL source=git@github.com:kyma-incubator/e2e-provisioning.git

ENTRYPOINT test.test
