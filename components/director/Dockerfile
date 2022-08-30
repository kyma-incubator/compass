FROM golang:1.18.5-alpine3.16 as director-builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . .

RUN go build -v -o director ./cmd/director/main.go \
  && mkdir /app && mv ./director /app/director

FROM golang:1.18.5-alpine3.16 as ns-adapter-builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . .

RUN go build -v -o ns-adapter ./cmd/ns-adapter/main.go \
  && mkdir /app && mv ./ns-adapter /app/ns-adapter

FROM golang:1.18.5-alpine3.16 as tenantfetcher-svc-builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . .

RUN go build -v -o tenantfetcher-svc ./cmd/tenantfetcher-svc/main.go \
  && mkdir /app && mv ./tenantfetcher-svc /app/tenantfetcher-svc

FROM golang:1.18.5-alpine3.16 as tenantloader-builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . .

RUN go build -v -o tenantloader ./cmd/tenantloader/main.go \
  && mkdir /app && mv ./tenantloader /app/tenantloader

FROM golang:1.18.5-alpine3.16 as ordaggregator-builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . .

RUN go build -v -o ordaggregator ./cmd/ordaggregator/main.go \
  && mkdir /app && mv ./ordaggregator /app/ordaggregator

FROM golang:1.18.5-alpine3.16 as scopessynchronizer-builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . .

RUN go build -v -o scopessynchronizer ./cmd/scopessynchronizer/main.go \
  && mkdir /app && mv mv ./scopessynchronizer /app/scopessynchronizer

FROM golang:1.18.5-alpine3.16 as systemfetcher-builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/director
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY . .

RUN go build -v -o systemfetcher ./cmd/systemfetcher/main.go \
  && mkdir /app && mv mv ./systemfetcher /app/systemfetcher

FROM alpine:3.16.2
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

#
# Copy binary
#

RUN apk --no-cache add curl ca-certificates

COPY --from=director-builder /app/director /app/director
COPY --from=ns-adapter-builder /app/ns-adapter /app/ns-adapter
COPY --from=tenantfetcher-builder /app/tenantfetcher /app/tenantfetcher
COPY --from=tenantloader-builder /app/tenantloader /app/tenantloader
COPY --from=ordaggregator-builder /app/ordaggregator /app/ordaggregator
COPY --from=scopessynchronizer-builder /app/scopessynchronizer /app/scopessynchronizer
COPY --from=systemfetcher-builder /app/systemfetcher /app/systemfetcher
COPY ./examples/ /app/examples/
COPY ./licenses ./app/licenses

#
# Run app
#

CMD ["/app/director"]
