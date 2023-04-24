FROM --platform=$BUILDPLATFORM golang:1.20-alpine3.17 as builder

ENV BASE_APP_DIR /go/src/github.com/kyma-incubator/compass/components/ias-broker
WORKDIR ${BASE_APP_DIR}

COPY go.mod go.sum ${BASE_APP_DIR}/
RUN go mod download -x

COPY cmd ${BASE_APP_DIR}/cmd
COPY internal ./internal

ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /main cmd/adapter/main.go

FROM alpine:3.17.3 as runtime
LABEL source = git@github.com:kyma-incubator/compass.git
WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /main /app/

RUN adduser --system --no-create-home --uid 1000 robot

USER 1000

EXPOSE 8080

CMD [ "/app/main" ]