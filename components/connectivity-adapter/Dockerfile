FROM golang:1.13.8-alpine3.11 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/connectivity-adapter
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#
COPY . ${BASE_APP_DIR}

#
# Build app
#

RUN go build -v -o main ./cmd/main.go
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses

FROM alpine:3.11.3
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
