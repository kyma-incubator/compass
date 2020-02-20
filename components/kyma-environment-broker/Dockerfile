# Build image
FROM golang:1.13.8-alpine3.11 AS build

WORKDIR /go/src/github.com/kyma-incubator/compass/components/kyma-environment-broker

COPY cmd cmd
COPY internal internal
COPY vendor vendor

RUN CGO_ENABLED=0 go build -o /bin/kyma-env-broker ./cmd/broker/main.go

# Get latest CA certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# Final image
FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/kyma-env-broker /bin/kyma-env-broker

CMD ["/bin/kyma-env-broker"]
