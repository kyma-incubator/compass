FROM golang:1.12.5-alpine3.9 as builder

ENV SRC_DIR=/go/src/github.com/kyma-incubator/compass/tests/e2e/provisioning

WORKDIR $SRC_DIR
COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test

# contains bash and kubectl
FROM linkyard/kubectl:1.14.2

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/e2e/provisioning/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/e2e/provisioning/test.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/e2e/provisioning/licenses ./licenses

VOLUME configs

LABEL source=git@github.com:kyma-incubator/e2e-provisioning.git

ENTRYPOINT ./entrypoint.sh
