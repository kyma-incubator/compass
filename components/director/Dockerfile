FROM golang:1.13.5-alpine3.10 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

#
# Copy files
#

COPY . .
# Build app
#

RUN go build -v -o director ./cmd/director/main.go && go build -v -o tenantloader ./cmd/tenantloader/main.go
RUN mkdir /app && mv ./director /app/director  && mv ./tenantloader /app/tenantloader # && mv ./licenses /app/licenses

FROM alpine:3.10
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

#
# Copy binary
#
RUN apk --no-cache add curl

COPY --from=builder /app /app
COPY ./examples/ /app/examples/

#
# Run app
#

CMD ["/app/director"]
