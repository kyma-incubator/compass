FROM golang:1.18.0-alpine3.15 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/external-services-mock
WORKDIR ${BASE_APP_DIR}

#
# Download dependencies
#

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

#
# Copy files
#

COPY . .

#
# Build app
#

RUN go build -v -o main ./cmd/main.go
RUN mkdir /app && mv ./main /app/main

FROM alpine:3.15.3
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
