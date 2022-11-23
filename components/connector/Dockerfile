FROM golang:1.19.3-alpine3.16 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/connector
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . ${BASE_APP_DIR}

RUN go build -v -o main ./cmd
RUN mkdir /app && mv ./main /app/main



FROM alpine:3.16.2
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

COPY --from=builder /app /app

CMD ["/app/main"]
