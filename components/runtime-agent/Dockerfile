FROM golang:1.12.5-alpine3.9 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/runtime-agent
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY ./vendor/ ${BASE_APP_DIR}/vendor/
COPY ./cmd/ ${BASE_APP_DIR}/cmd/
COPY ./licenses ${BASE_APP_DIR}/licenses

#
# Build app
#

RUN go build -v -o runtime-agent ./cmd/main.go
RUN mkdir /app && mv ./runtime-agent /app/runtime-agent && mv ./licenses /app/licenses

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

CMD ["/app/runtime-agent"]
