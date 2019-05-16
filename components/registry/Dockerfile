FROM golang:1.12.5-alpine3.9 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/registry
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./cmd/main.go ${BASE_APP_DIR}/

#
# Build app
#

RUN go build -v -o main .
RUN mkdir /app && mv ./main /app/main

FROM alpine:3.9
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

#
# Copy binary
#

COPY --from=builder /app /app

#
# Run app
#

CMD ["/app/main"]
