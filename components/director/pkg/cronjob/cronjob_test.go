package cronjob_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/stretchr/testify/assert"
)

func TestCronJob(t *testing.T) {
	const (
		maxCronJobRuns        = 10 // Used for safeguarding the tests in case of infinite loop
		defaultSchedulePeriod = time.Nanosecond
	)

	testCases := []struct {
		Name                string
		FnBody              func(executionsCount int, cancel context.CancelFunc)
		ExpectedCronJobRuns int
		SchedulePeriod      time.Duration
	}{
		{
			Name:                "Should run cronJob until the context end",
			ExpectedCronJobRuns: 3,
			FnBody: func(executionsCount int, cancelCtx context.CancelFunc) {
				if executionsCount == 3 {
					cancelCtx()
				}
			},
			SchedulePeriod: defaultSchedulePeriod,
		},
		{
			Name: "Should not schedule next cronJob in parallel if execution takes more than the wait period",

			ExpectedCronJobRuns: 1,
			FnBody: func(executionsCount int, cancelCtx context.CancelFunc) {
				<-time.After(defaultSchedulePeriod * 2)
				cancelCtx()
			},
			SchedulePeriod: defaultSchedulePeriod,
		},
		{
			Name:                "Should schedule next cronJob if execution takes more than the wait period",
			ExpectedCronJobRuns: 3,
			FnBody: func(executionsCount int, cancelCtx context.CancelFunc) {
				<-time.After(defaultSchedulePeriod * 2)
				if executionsCount == 3 {
					cancelCtx()
				}
			},
			SchedulePeriod: defaultSchedulePeriod,
		},
		{
			Name:                "Should stop schedule immediately if context is canceled during waiting",
			ExpectedCronJobRuns: 1,
			FnBody: func(executionsCount int, cancelCtx context.CancelFunc) {
				// give some time for the function to exit and the CronJob to start waiting for the next execution
				time.AfterFunc(time.Millisecond*5, func() {
					cancelCtx()
				})
			},
			SchedulePeriod: time.Minute,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			cronJob := cronjob.CronJob{
				Name:           testCase.Name,
				SchedulePeriod: testCase.SchedulePeriod,
			}
			cronJobRuns := 0
			ctx, cancel := context.WithCancel(context.Background())

			cronJob.Fn = func(ctx context.Context) {
				cronJobRuns += 1
				if cronJobRuns == maxCronJobRuns {
					cancel()
				} else {
					testCase.FnBody(cronJobRuns, cancel)
				}
			}
			err := cronjob.RunCronJob(ctx, cronjob.ElectionConfig{ElectionEnabled: false}, cronJob)
			assert.NoError(t, err)

			assert.Equal(t, cronJobRuns, testCase.ExpectedCronJobRuns)
		})
	}
}
