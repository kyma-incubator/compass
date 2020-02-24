FROM golang:1.13.8-alpine3.11 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY . .

#
# Build app
#

RUN go build -v -o director ./cmd/director/main.go \
  && go build -v -o tenantfetcher ./cmd/tenantfetcher/main.go \
  && go build -v -o tenantloader ./cmd/tenantloader/main.go
RUN mkdir /app && mv ./director /app/director \
  && mv ./tenantfetcher /app/tenantfetcher \
  && mv ./tenantloader /app/tenantloader \
  && mv ./licenses /app/licenses

FROM alpine:3.11.3
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
