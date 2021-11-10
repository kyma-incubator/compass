FROM golang:1.17.0-alpine3.14 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
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

RUN go build -v -o ordaggregator ./cmd/ordaggregator/main.go
RUN mkdir /app && mv ./ordaggregator /app/ordaggregator \
  && mv ./licenses /app/licenses

FROM alpine:3.14.2
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

#
# Copy binary
#

RUN apk --no-cache add curl ca-certificates

COPY --from=builder /app /app
COPY ./examples/ /app/examples/

#
# Run app
#

CMD ["/app/director"]
