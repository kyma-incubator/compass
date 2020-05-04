FROM golang:1.14.2-alpine3.11 as build

ARG VERSION
ARG CGO_ENABLED
ARG GO111MODULE

ENV VERSION=${VERSION:-dev}
ENV CGO_ENABLED=${CGO_ENABLED:-0}
ENV GO111MODULE=${GO111MODULE:-on}
ENV BASE_APP_DIR=/go/src/github.com/kyma-incubator/compass/components/metris

WORKDIR ${BASE_APP_DIR}

COPY . .

# RUN go mod download
RUN go build -v -o /bin/metris -ldflags="-s -w -X main.version=${VERSION}" ./cmd/metris/main.go

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
LABEL source=git@github.com:kyma-incubator/compass.git

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bin/metris /bin/metris

ENTRYPOINT ["/bin/metris"]