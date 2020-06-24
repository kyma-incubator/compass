FROM golang:1.14.4-alpine3.12 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/connector
WORKDIR ${BASE_APP_DIR}

COPY ./internal/ ${BASE_APP_DIR}/internal/
COPY ./config/ ${BASE_APP_DIR}/config/
COPY ./pkg/ ${BASE_APP_DIR}/pkg/
COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./cmd/main.go ${BASE_APP_DIR}/
COPY ./licenses ${BASE_APP_DIR}/licenses

RUN apk add -U --no-cache ca-certificates && update-ca-certificates

RUN go build -v -o main .
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses



FROM scratch
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

COPY --from=builder /app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/main"]
