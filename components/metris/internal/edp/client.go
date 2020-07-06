package edp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/kyma-project/control-plane/components/metris/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"k8s.io/client-go/util/workqueue"
)

// type Event []byte

type Config struct {
	URL               string        `kong:"help='EDP base URL',env='EDP_URL',default='https://input.yevents.io',required=true"`
	Token             string        `kong:"help='EDP source token',placeholder='SECRET',env='EDP_TOKEN',required=true"`
	Namespace         string        `kong:"help='EDP Namespace',env='EDP_NAMESPACE',required=true"`
	DataStream        string        `kong:"help='EDP data stream name',env='EDP_DATASTREAM_NAME',required=true"`
	DataStreamVersion string        `kong:"help='EDP data stream version',env='EDP_DATASTREAM_VERSION',required=true"`
	DataStreamEnv     string        `kong:"help='EDP data stream environment',env='EDP_DATASTREAM_ENV',required=true"`
	Timeout           time.Duration `kong:"help='Time limit for requests made by the EDP client',env='EDP_TIMEOUT',required=true,default='30s'"`
	Buffer            int           `kong:"help='Number of events that the buffer can have.',env='EDP_BUFFER',required=true,default=100"`
	Workers           int           `kong:"help='Number of workers to send metrics.',env='EDP_WORKERS',required=true,default=5"`
	EventRetry        int           `kong:"help='Number of retries for sending event.',env='EDP_RETRY',required=true,default=5"`
}

type Client struct {
	config        *Config
	httpClient    *http.Client
	logger        *zap.SugaredLogger
	queue         workqueue.RateLimitingInterface
	eventsChannel <-chan *[]byte
}

var (
	ErrEventInvalidRequest    = errors.New("invalid request")
	ErrEventMissingParameters = errors.New("namespace, dataStream or dataTenant not found")
	ErrEventUnknown           = errors.New("unknown error")
	ErrEventUnmarshal         = errors.New("unmarshal error")
	ErrEventHTTPRequest       = errors.New("HTTP request error")

	rateLimiterBaseDelay      = 5 * time.Second
	rateLimiterMaxDelay       = 60 * time.Second
	clientReqTimeout          = 30 * time.Second
	clientIdleConnTimeout     = 60 * time.Second
	clientTLSHandshakeTimeout = 10 * time.Second

	defaultHTTPClient = &http.Client{
		Timeout: clientReqTimeout,
		Transport: &http.Transport{
			IdleConnTimeout:     clientIdleConnTimeout,
			TLSHandshakeTimeout: clientTLSHandshakeTimeout,
		},
	}
)

func NewClient(c *Config, httpClient *http.Client, eventsChannel <-chan *[]byte, logger *zap.SugaredLogger) *Client {
	if httpClient == nil {
		httpClient = defaultHTTPClient
		httpClient.Timeout = c.Timeout
	}

	// retry after baseDelay*2^<num-failures>
	ratelimiter := workqueue.NewItemExponentialFailureRateLimiter(rateLimiterBaseDelay, rateLimiterMaxDelay)

	return &Client{
		config:        c,
		httpClient:    httpClient,
		queue:         workqueue.NewNamedRateLimitingQueue(ratelimiter, "edp-events"),
		logger:        logger.With("component", "edp"),
		eventsChannel: eventsChannel,
	}
}

func (c *Client) Run(parentCtx context.Context, parentwg *sync.WaitGroup) {
	c.logger.Debug("starting ingester")

	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	parentwg.Add(1)
	defer parentwg.Done()

	var wg sync.WaitGroup

	wg.Add(c.config.Workers)

	for i := 0; i < c.config.Workers; i++ {
		go func(i int) {
			defer wg.Done()

			workerlogger := c.logger.With("worker", i)

			for {
				event, quit := c.queue.Get()
				if quit {
					workerlogger.Debugf("shutting down")
					return
				}

				c.handleErr(
					c.Write(ctx, event.(*[]byte), workerlogger),
					event.(*[]byte), workerlogger,
				)

				c.queue.Done(event)
			}
		}(i)
	}

	go func() {
		for {
			select {
			case event := <-c.eventsChannel:
				c.queue.Add(event)
			case <-ctx.Done():
				c.queue.ShutDown()
				return
			}
		}
	}()

	wg.Wait()
	c.logger.Debug("stopping ingester")
}

// handleErr checks if an error happened and requeue to retry sending the event
func (c *Client) handleErr(err error, event *[]byte, logger *zap.SugaredLogger) {
	if err == nil {
		// if no error, clear number of queue history
		c.queue.Forget(event)
		return
	}

	// if the error is an unmarshall one, we remove it from the queue
	if errors.Is(err, ErrEventUnmarshal) {
		logger.Error(err)
		c.queue.Done(event)

		return
	}

	failures := c.queue.NumRequeues(event)

	// retries X times, then stops trying
	if failures < c.config.EventRetry {
		nextsend := when(failures)
		logger.With(
			"error", err,
			"event", string(*event),
		).Warnf("error sending event, requeuing in %s (%d/%d)", nextsend, failures, c.config.EventRetry)

		// Re-enqueue the event
		c.queue.AddRateLimited(event)

		return
	}

	c.queue.Forget(event)
	logger.With(
		"error", err,
		"event", string(*event),
	).Errorf("failed %d times to send the event, removing it out of the queue", c.config.EventRetry)
}

// Write sends a batch of samples(json) to EDP
func (c *Client) Write(parentctx context.Context, data *[]byte, logger *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(parentctx)
	defer cancel()

	// try to unmarshal json data submitted
	var events map[string]json.RawMessage

	err := json.Unmarshal(*data, &events)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrEventUnmarshal, err)
	}

	// datatenant == subaccountid
	for datatenant, event := range events {
		metricTimer := prometheus.NewTimer(metrics.SentSamplesDuration)
		defer metricTimer.ObserveDuration()

		receivedEvents := 1

		eventjson, err := event.MarshalJSON()
		if err == nil {
			var datamap []interface{}

			err = json.Unmarshal(eventjson, &datamap)
			if err == nil {
				receivedEvents = len(datamap)
			}
		}

		edpurl := fmt.Sprintf("%s/namespaces/%s/dataStreams/%s/%s/dataTenants/%s/%s/eventBatch",
			c.config.URL,
			c.config.Namespace,
			c.config.DataStream,
			c.config.DataStreamVersion,
			datatenant,
			c.config.DataStreamEnv,
		)

		logger.Debugf("sending events '%s':\n%+v", edpurl, string(event))

		httpreq, err := http.NewRequestWithContext(ctx, "POST", edpurl, bytes.NewBuffer(event))
		if err != nil {
			metrics.FailedSamples.Add(float64(receivedEvents))
			return fmt.Errorf("%w: %s", ErrEventHTTPRequest, err)
		}

		httpreq.Header.Set("User-Agent", "metris")
		httpreq.Header.Add("Content-Type", "application/json;charset=utf-8")
		httpreq.Header.Add("Authorization", "bearer "+c.config.Token)

		resp, err := c.httpClient.Do(httpreq)
		if err != nil {
			metrics.FailedSamples.Add(float64(receivedEvents))
			return fmt.Errorf("%w: %s", ErrEventHTTPRequest, err)
		}

		defer func() {
			err := resp.Body.Close()
			if err != nil {
				c.logger.Warn(err)
			}
		}()

		if resp.StatusCode != http.StatusCreated {
			metrics.FailedSamples.Add(float64(receivedEvents))
			return statusError(resp.StatusCode)
		}

		metrics.SentSamples.Add(float64(receivedEvents))
	}

	return nil
}

func statusError(code int) error {
	var err error

	switch code {
	case http.StatusBadRequest:
		err = ErrEventInvalidRequest
	case http.StatusNotFound:
		err = ErrEventMissingParameters
	default:
		err = ErrEventUnknown
	}

	return fmt.Errorf("%w: %d", err, code)
}

func when(failure int) time.Duration {
	var rateLimiterBase float64 = 2

	backoff := float64(rateLimiterBaseDelay.Nanoseconds()) * math.Pow(rateLimiterBase, float64(failure))
	if backoff > math.MaxInt64 {
		return rateLimiterMaxDelay
	}

	when := time.Duration(backoff)
	if when > rateLimiterMaxDelay {
		return rateLimiterMaxDelay
	}

	return when
}
