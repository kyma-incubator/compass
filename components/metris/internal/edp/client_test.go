package edp_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/metris/internal/edp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	defaultLogger = zap.NewNop().Sugar()
	defaultconfig = &edp.Config{
		URL:               "http://127.0.0.1:9999",
		Token:             "E6B99A13-783F-4A3B-8605-C5EA32CA44B5",
		Timeout:           30 * time.Second,
		Namespace:         "kyma-dev",
		DataStream:        "consumption-metrics",
		DataStreamVersion: "1",
		DataStreamEnv:     "dev",
		Buffer:            100,
		Workers:           1,
		EventRetry:        1,
	}
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func fakeTestClient(status int, err error) *http.Client {
	var fn roundTripFunc = func(req *http.Request) (*http.Response, error) {
		if err != nil {
			return nil, err
		}

		return &http.Response{StatusCode: status, Body: ioutil.NopCloser(bytes.NewBufferString("")), Header: make(http.Header)}, nil
	}

	return &http.Client{
		Transport: fn,
	}
}

func TestClient_CreateEventSucess(t *testing.T) {
	fakehttpclient := fakeTestClient(http.StatusCreated, nil)
	client := edp.NewClient(defaultconfig, fakehttpclient, nil, defaultLogger)

	data := []byte(`{"event":[{"data":"test"}]}`)

	err := client.Write(context.TODO(), &data, defaultLogger)
	assert.NoError(t, err)
}

func TestClient_CreateEventInvalid(t *testing.T) {
	fakehttpclient := fakeTestClient(http.StatusBadRequest, nil)
	client := edp.NewClient(defaultconfig, fakehttpclient, nil, defaultLogger)

	data := []byte(`{"event":[{"data":"test2"}]}`)

	err := client.Write(context.TODO(), &data, defaultLogger)
	assert.Conditionf(t, func() bool {
		return errors.Is(err, edp.ErrEventInvalidRequest)
	}, "invalid error, got %s, should be: %s", err, edp.ErrEventInvalidRequest)
}

func TestClient_CreateEventMissingParam(t *testing.T) {
	fakehttpclient := fakeTestClient(http.StatusNotFound, nil)
	client := edp.NewClient(defaultconfig, fakehttpclient, nil, defaultLogger)

	data := []byte(`{"test3":[{"error":""}]}`)

	err := client.Write(context.TODO(), &data, defaultLogger)
	assert.Conditionf(t, func() bool {
		return errors.Is(err, edp.ErrEventMissingParameters)
	}, "invalid error, got %s, should be: %s", err, edp.ErrEventMissingParameters)
}

func TestClient_CreateEventUnknownError(t *testing.T) {
	fakehttpclient := fakeTestClient(http.StatusUnauthorized, nil)
	client := edp.NewClient(defaultconfig, fakehttpclient, nil, defaultLogger)

	data := []byte(`{"test4":[{"error":""},{"error":""}]}`)

	err := client.Write(context.TODO(), &data, defaultLogger)
	assert.Conditionf(t, func() bool {
		return errors.Is(err, edp.ErrEventUnknown)
	}, "invalid error, got %s, should be: %s", err, edp.ErrEventUnknown)
}

func TestClient_CreateEventJSONError(t *testing.T) {
	fakehttpclient := fakeTestClient(http.StatusCreated, nil)
	client := edp.NewClient(defaultconfig, fakehttpclient, nil, defaultLogger)

	data := []byte(`{"test5":[{"error":""}]`)

	err := client.Write(context.TODO(), &data, defaultLogger)
	assert.Conditionf(t, func() bool {
		return errors.Is(err, edp.ErrEventUnmarshal)
	}, "invalid error, got %s, should be: %s", err, edp.ErrEventUnmarshal)
}

func TestClient_CreateEventHTTPError(t *testing.T) {
	fakehttpclient := fakeTestClient(0, fmt.Errorf("network error"))
	client := edp.NewClient(defaultconfig, fakehttpclient, nil, defaultLogger)

	data := []byte(`{"test6":[{"error":""}]}`)

	err := client.Write(context.TODO(), &data, defaultLogger)
	assert.Conditionf(t, func() bool {
		return errors.Is(err, edp.ErrEventHTTPRequest)
	}, "invalid error, got %s, should be: %s", err, edp.ErrEventHTTPRequest)
}

// func TestListen(t *testing.T) {
// 	fakehttpclient := fakeTestClient(http.StatusCreated, nil)

// 	eventsChannel := make(chan *[]byte, 1)
// 	defer close(eventsChannel)
// 	client := edp.NewClient(defaultconfig, fakehttpclient, eventsChannel, zap.NewNop().Sugar())

// 	ctx, cancel := context.WithCancel(context.TODO())
// 	wg := sync.WaitGroup{}

// 	go client.Run(ctx, &wg)

// 	data := []byte(`{"test":["status":"ok"]}`)
// 	eventsChannel <- &data

// 	time.Sleep(1 * time.Second)

// 	cancel()
// 	wg.Wait()
// }
