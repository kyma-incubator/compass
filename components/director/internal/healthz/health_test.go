package healthz_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/healthz"
	"github.com/kyma-incubator/compass/components/director/internal/healthz/automock"
	"github.com/stretchr/testify/require"
)

var (
	firstConfig = healthz.IndicatorConfig{
		Name:         "First",
		Interval:     time.Second,
		Timeout:      time.Second,
		InitialDelay: time.Second,
		Threshold:    1,
	}

	secondConfig = healthz.IndicatorConfig{
		Name:         "Second",
		Interval:     time.Second,
		Timeout:      time.Second,
		InitialDelay: time.Second,
		Threshold:    1,
	}
)

func TestNew(t *testing.T) {
	t.Run("should return not nil health", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		// WHEN
		health, err := healthz.New(ctx, healthz.Config{})
		// THEN
		require.NotNil(t, health)
		require.NoError(t, err)
	})
	t.Run("should return error on invalid interval", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		// WHEN
		health, err := healthz.New(ctx, healthz.Config{Indicators: []healthz.IndicatorConfig{{
			Interval: 0,
		}}})
		// THEN
		require.Nil(t, health)
		require.Error(t, err)
		require.Contains(t, err.Error(), "interval")
	})
	t.Run("should return error on invalid timeout", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		// WHEN
		health, err := healthz.New(ctx, healthz.Config{Indicators: []healthz.IndicatorConfig{{
			Interval: time.Second,
			Timeout:  0,
		}}})
		// THEN
		require.Nil(t, health)
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout")
	})
	t.Run("should return error on invalid initial delay", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		// WHEN
		health, err := healthz.New(ctx, healthz.Config{Indicators: []healthz.IndicatorConfig{{
			Interval:     time.Second,
			Timeout:      time.Second,
			InitialDelay: -1,
		}}})
		// THEN
		require.Nil(t, health)
		require.Error(t, err)
		require.Contains(t, err.Error(), "initial delay ")
	})
	t.Run("should return error on invalid threshold", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		// WHEN
		health, err := healthz.New(ctx, healthz.Config{Indicators: []healthz.IndicatorConfig{{
			Interval:     time.Second,
			Timeout:      time.Second,
			InitialDelay: time.Second,
			Threshold:    -1,
		}}})
		// THEN
		require.Nil(t, health)
		require.Error(t, err)
		require.Contains(t, err.Error(), "threshold")
	})
}

func TestFullFlow(t *testing.T) {
	t.Run("should configure properly(first config exists) and return UP when both indicators succeed", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		healthCfg := healthz.Config{Indicators: []healthz.IndicatorConfig{firstConfig}}

		firstStatus := &automock.Status{}
		defer firstStatus.AssertExpectations(t)
		firstStatus.On("Error").Return(nil)

		firstInd := &automock.Indicator{}
		defer firstInd.AssertExpectations(t)
		firstInd.On("Name").Return("First")
		firstInd.On("Run", ctx).Return(nil)
		firstInd.On("Status").Return(firstStatus)
		firstInd.On("Configure", firstConfig).Return()

		secondStatus := &automock.Status{}
		defer secondStatus.AssertExpectations(t)
		secondStatus.On("Error").Return(nil)

		secondInd := &automock.Indicator{}
		defer secondInd.AssertExpectations(t)
		secondInd.On("Name").Return("Second")
		secondInd.On("Run", ctx).Return(nil)
		secondInd.On("Status").Return(secondStatus)
		secondInd.On("Configure", healthz.NewDefaultConfig()).Return()

		// WHEN
		health, err := healthz.New(ctx, healthCfg)
		require.NoError(t, err)

		health.RegisterIndicator(firstInd).
			RegisterIndicator(secondInd).
			Start()
		status := health.ReportStatus()

		// THEN
		require.Equal(t, status, healthz.UP)
		AssertHandlerStatusCodeForHealth(t, health, http.StatusOK, healthz.UP)
	})

	t.Run("should configure properly(both config exist) and return DOWN when one indicator fails", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		healthCfg := healthz.Config{Indicators: []healthz.IndicatorConfig{firstConfig, secondConfig}}

		firstStatus := &automock.Status{}
		defer firstStatus.AssertExpectations(t)
		firstStatus.On("Error").Return(errors.New("some error"))
		firstStatus.On("Details").Return("some details")

		firstInd := &automock.Indicator{}
		defer firstInd.AssertExpectations(t)
		firstInd.On("Name").Return("First").Times(3)
		firstInd.On("Run", ctx).Return(nil)
		firstInd.On("Status").Return(firstStatus)
		firstInd.On("Configure", firstConfig).Return()

		secondStatus := &automock.Status{}
		defer secondStatus.AssertExpectations(t)
		secondStatus.On("Error").Return(nil)

		secondInd := &automock.Indicator{}
		defer secondInd.AssertExpectations(t)
		secondInd.On("Name").Return("Second")
		secondInd.On("Run", ctx).Return(nil)
		secondInd.On("Status").Return(secondStatus)
		secondInd.On("Configure", secondConfig).Return()

		// WHEN
		health, err := healthz.New(ctx, healthCfg)
		require.NoError(t, err)

		health.RegisterIndicator(firstInd).
			RegisterIndicator(secondInd).
			Start()

		status := health.ReportStatus()

		// THEN
		require.Equal(t, status, healthz.DOWN)
		AssertHandlerStatusCodeForHealth(t, health, http.StatusInternalServerError, healthz.DOWN)
	})

	t.Run("should configure properly(neither config exist) and return DOWN when all indicators fail", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		healthCfg := healthz.Config{Indicators: []healthz.IndicatorConfig{}}

		firstStatus := &automock.Status{}
		defer firstStatus.AssertExpectations(t)
		firstStatus.On("Error").Return(errors.New("some error"))
		firstStatus.On("Details").Return("some details")

		firstInd := &automock.Indicator{}
		defer firstInd.AssertExpectations(t)
		firstInd.On("Name").Return("First").Times(3)
		firstInd.On("Run", ctx).Return(nil)
		firstInd.On("Status").Return(firstStatus)
		firstInd.On("Configure", healthz.NewDefaultConfig()).Return()

		secondStatus := &automock.Status{}
		defer secondStatus.AssertExpectations(t)
		secondStatus.On("Error").Return(errors.New("some error 2"))
		secondStatus.On("Details").Return("some details 2")

		secondInd := &automock.Indicator{}
		defer secondInd.AssertExpectations(t)
		secondInd.On("Name").Return("Second").Times(3)
		secondInd.On("Run", ctx).Return(nil)
		secondInd.On("Status").Return(secondStatus)
		secondInd.On("Configure", healthz.NewDefaultConfig()).Return()

		// WHEN
		health, err := healthz.New(ctx, healthCfg)
		require.NoError(t, err)

		health.RegisterIndicator(firstInd).
			RegisterIndicator(secondInd).
			Start()

		status := health.ReportStatus()

		// THEN
		require.Equal(t, status, healthz.DOWN)
		AssertHandlerStatusCodeForHealth(t, health, http.StatusInternalServerError, healthz.DOWN)
	})
}

func AssertHandlerStatusCodeForHealth(t *testing.T, h *healthz.Health, expectedCode int, expectedBody string) {
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthz.NewHealthHandler(h))
	// WHEN
	handler.ServeHTTP(rr, req)
	// THEN
	require.Equal(t, expectedCode, rr.Code)
	require.Equal(t, expectedBody, rr.Body.String())
}
