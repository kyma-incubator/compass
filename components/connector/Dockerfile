FROM golang:1.18.2-alpine3.15 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/connector
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . ${BASE_APP_DIR}

RUN go build -v -o main ./cmd
RUN mkdir /app && mv ./main /app/main && mv ./licenses /app/licenses



FROM alpine:3.15.3
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

COPY --from=builder /app /app

CMD ["/app/main"]
