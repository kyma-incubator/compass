FROM golang:1.20-alpine3.17 as builder

WORKDIR /go/src/github.com/kyma-incubator/compass/components/ias-broker

COPY cmd ./cmd
COPY internal ./internal
COPY go.mod go.sum ./

RUN go build -o /main cmd/adapter/main.go

FROM alpine:3.17.2 as runtime

WORKDIR /app

COPY --from=builder /main /app/

RUN adduser --system --no-create-home --uid 1000 robot

USER 1000

EXPOSE 8080

ENTRYPOINT [ "./main" ]