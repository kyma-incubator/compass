FROM golang:1.14.4-alpine3.12 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/external-services-mock
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY . .

#
# Build app
#

RUN go build -v -o main ./cmd/main.go
RUN mkdir /app && mv ./main /app/main

FROM alpine:3.12.0
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
