FROM golang:1.15.5-alpine3.12 as builder

ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests/async-operations

WORKDIR ${BASE_TEST_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./tests -o async-operations.test

FROM alpine:3.12.0

RUN apk add --no-cache curl

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/async-operations/async-operations.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/async-operations/licenses ./licenses