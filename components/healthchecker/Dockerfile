FROM golang:1.15.7-alpine3.12 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/healthchecker
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./cmd/main.go ${BASE_APP_DIR}/
COPY ./licenses ${BASE_APP_DIR}/licenses

#
# Build app
#

RUN go build -v -o main .
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses

FROM alpine:3.12.3
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
