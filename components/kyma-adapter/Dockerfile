FROM --platform=$BUILDPLATFORM golang:1.21.6-alpine3.19 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/kyma-adapter
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

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -v -o main ./cmd/main.go
RUN mkdir /app && mv ./main /app/main

FROM alpine:3.19
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
