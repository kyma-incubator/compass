package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/event"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"
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
	assert.NoError(t, wait.PollImmediate(20*time.Millisecond, 2*time.Second, func() (bool, error) {
		return containsA(gotEventAList1, eventA{msg: "first event"}) &&
			containsA(gotEventAList1, eventA{msg: "third event"}) &&
			containsA(gotEventAList2, eventA{msg: "first event"}) &&
			containsA(gotEventAList2, eventA{msg: "third event"}) &&
			containsB(gotEventBList, eventB{msg: "second event"}), nil
	}))
}

func containsA(slice []eventA, item eventA) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsB(slice []eventB, item eventB) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type eventA struct {
	msg string
}

type eventB struct {
	msg string
}
