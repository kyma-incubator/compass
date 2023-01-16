FROM --platform=$BUILDPLATFORM golang:1.19.3-alpine3.16 as builder

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

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -v -o director ./cmd/director/main.go \
  && go build -v -o ns-adapter ./cmd/ns-adapter/main.go \
  && go build -v -o tenantfetcher-svc ./cmd/tenantfetcher-svc/main.go \
  && go build -v -o tenantloader ./cmd/tenantloader/main.go \
  && go build -v -o ordaggregator ./cmd/ordaggregator/main.go \
  && go build -v -o scopessynchronizer ./cmd/scopessynchronizer/main.go \
  && go build -v -o systemfetcher ./cmd/systemfetcher/main.go \
  && go build -v -o destinationfetcher ./cmd/destinationfetcher-svc/main.go
RUN mkdir /app && mv ./director /app/director \
  && mv ./ns-adapter /app/ns-adapter \
  && mv ./tenantfetcher-svc /app/tenantfetcher-svc \
  && mv ./tenantloader /app/tenantloader \
  && mv ./ordaggregator /app/ordaggregator \
  && mv ./scopessynchronizer /app/scopessynchronizer \
  && mv ./systemfetcher /app/systemfetcher \
  && mv ./destinationfetcher /app/destinationfetcher

FROM --platform=$BUILDPLATFORM alpine:3.16.2
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
