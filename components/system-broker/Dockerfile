FROM golang:1.14.4-alpine3.12 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/system-broker
WORKDIR ${BASE_APP_DIR}

COPY . .

RUN go build -v -o system-broker ./cmd/main.go

FROM alpine:edge
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

RUN apk --no-cache add curl ca-certificates

COPY --from=builder /app /app

CMD ["/app/system-broker"]
