FROM golang:1.15.5-alpine3.12 as builder

ENV SRC_DIR=/go/src/github.com/kyma-incubator/compass/tests/connector-tests

WORKDIR $SRC_DIR

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOOS=linux go test -c ./test/apitests



FROM alpine:3.12.3

LABEL source=git@github.com:kyma-project/kyma.git

RUN apk --no-cache add ca-certificates curl

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/apitests.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connector-tests/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
