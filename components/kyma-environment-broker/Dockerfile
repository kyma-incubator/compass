# Build image
FROM golang:1.14.4-alpine3.12 AS build

WORKDIR /go/src/github.com/kyma-incubator/compass/components/kyma-environment-broker

COPY cmd cmd
COPY internal internal
COPY vendor vendor

RUN CGO_ENABLED=0 go build -o /bin/kyma-env-broker ./cmd/broker/main.go \
	&& CGO_ENABLED=0 go build -o /bin/environments-cleanup ./cmd/environmentscleanup/main.go

# Get latest CA certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# Final image
FROM alpine:edge

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/kyma-env-broker /bin/kyma-env-broker
COPY --from=build /bin/environments-cleanup /bin/environments-cleanup

CMD ["/bin/kyma-env-broker"]