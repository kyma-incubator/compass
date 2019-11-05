FROM golang:1.12.5-alpine3.9 as builder


ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests/end-to-end

WORKDIR ${BASE_TEST_DIR}
COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./director
RUN CGO_ENABLED=0 go test -c ./e2e

FROM alpine:3.10

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/end-to-end/wait-for-director.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/end-to-end/director.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/end-to-end/e2e.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/end-to-end/licenses ./licenses

CMD ["./director.test", "-test.v"]
