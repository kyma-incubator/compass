FROM golang:1.18.0-alpine3.15 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/pairing-adapter
WORKDIR ${BASE_APP_DIR}

#
# Download dependencies
#

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

#
# Copy files
#
COPY . ${BASE_APP_DIR}

#
# Build app
#

RUN go build -v -o main ./cmd/main.go
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses

FROM alpine:3.15.3
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

#
# Copy binary
#
RUN apk --no-cache add ca-certificates
COPY --from=builder /app /app

#
# Run app
#

CMD ["/app/main"]
