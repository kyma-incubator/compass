FROM golang:1.13.1-alpine3.10 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/provisioner
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./assets/ /assets/
COPY ./internal/ ${BASE_APP_DIR}/internal/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./cmd/ ${BASE_APP_DIR}/cmd
COPY ./licenses ${BASE_APP_DIR}/licenses

#
# Add ca-certificates
#

RUN apk add -U --no-cache ca-certificates && update-ca-certificates

#
# Build app
#

RUN go build -v -o main ./cmd/
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses

FROM alpine:3.10
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

#
# Copy binary
#

COPY --from=builder /assets /assets
COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

#
# Run app
#

CMD ["/app/main"]