FROM golang:1.13.8-alpine3.11 as builder


ENV BASE_TEST_DIR /go/src/github.com/kyma-incubator/compass/tests/connectivity-adapter

WORKDIR ${BASE_TEST_DIR}
COPY . ${BASE_TEST_DIR}

RUN CGO_ENABLED=0 go test -c ./test/api

# After the next release of alpine is available (3.12) we should set it here instead of the edge tag
# We can't do it now cause alpine:3.11 contains a vulnerable version of curl
FROM alpine:edge

RUN apk add --no-cache curl

LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connectivity-adapter/scripts/entrypoint.sh .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connectivity-adapter/api.test .
COPY --from=builder /go/src/github.com/kyma-incubator/compass/tests/connectivity-adapter/licenses ./licenses

ENTRYPOINT ./entrypoint.sh
