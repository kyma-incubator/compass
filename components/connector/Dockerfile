FROM golang:1.15.5-alpine3.12 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/connector
WORKDIR ${BASE_APP_DIR}

COPY . ${BASE_APP_DIR}

RUN go build -v -o main ./cmd
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses



FROM alpine:3.12.0
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

COPY --from=builder /app /app

CMD ["/app/main"]
