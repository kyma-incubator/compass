FROM eu.gcr.io/kyma-project/test-infra/buildpack-golang-toolbox:v20200117-d3885041 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-project/control-plane/components/kubeconfig-service
ENV CGO_ENABLED 0
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY . .

#
# Build app
#

RUN go build -v -o main ./cmd/generator/main.go
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses

FROM alpine:3.12.0
LABEL source = git@github.com:kyma-project/control-plane.git

WORKDIR /app

#
# Install certificates
#

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#
# Copy binary
#

COPY --from=builder /app /app

#
# Run app
#

CMD ["/app/main"]
