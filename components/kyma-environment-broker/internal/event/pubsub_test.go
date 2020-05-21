package event_test

import (
	"context"
	"testing"

	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"
	"github.com/stretchr/testify/assert"
)

func TestPubSub(t *testing.T) {
	// given
	var gotEventAList1 []eventA
	var gotEventAList2 []eventA
	handlerA1 := func(ctx context.Context, ev interface{}) error {
		gotEventAList1 = append(gotEventAList1, ev.(eventA))
		return nil
	}
	handlerA2 := func(ctx context.Context, ev interface{}) error {
		gotEventAList2 = append(gotEventAList2, ev.(eventA))
		return nil
	}
	var gotEventBList []eventB
	handlerB := func(ctx context.Context, ev interface{}) error {
		gotEventBList = append(gotEventBList, ev.(eventB))
		return nil
	}
	svc := event.NewPubSub()
	svc.Subscribe(eventA{}, handlerA1)
	svc.Subscribe(eventB{}, handlerB)
	svc.Subscribe(eventA{}, handlerA2)

	// when
	svc.Publish(context.TODO(), eventA{msg: "first event"})
	svc.Publish(context.TODO(), eventB{msg: "second event"})
	svc.Publish(context.TODO(), eventA{msg: "third event"})

	time.Sleep(1 * time.Millisecond)

	// then
	assert.NoError(t, wait.PollImmediate(20 * time.Millisecond, 2*time.Second, func() (bool, error) {
		return eventA{msg: "first event"} == gotEventAList1[0] &&
			eventA{msg: "first event"} == gotEventAList2[0] &&
			eventA{msg: "third event"} == gotEventAList1[1] &&
			eventA{msg: "third event"} == gotEventAList2[1] &&
			eventB{msg: "second event"} == gotEventBList[0], nil
	}))
}

type eventA struct {
	msg string
}

type eventB struct {
	msg string
}
