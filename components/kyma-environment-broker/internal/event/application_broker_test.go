package event_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event"
	"github.com/magiconair/properties/assert"
)

func TestNewApplicationEventBroker(t *testing.T) {
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
	svc := event.NewApplicationEventBroker()
	svc.Subscribe(eventA{}, handlerA1)
	svc.Subscribe(eventB{}, handlerB)
	svc.Subscribe(eventA{}, handlerA2)

	// when
	svc.Publish(context.TODO(), eventA{msg: "first event"})
	svc.Publish(context.TODO(), eventB{msg: "second event"})
	svc.Publish(context.TODO(), eventA{msg: "third event"})

	// then
	assert.Equal(t, eventA{msg: "first event"}, gotEventAList1[0])
	assert.Equal(t, eventA{msg: "first event"}, gotEventAList2[0])
	assert.Equal(t, eventA{msg: "third event"}, gotEventAList1[1])
	assert.Equal(t, eventA{msg: "third event"}, gotEventAList2[1])

	assert.Equal(t, eventB{msg: "second event"}, gotEventBList[0])
}

type eventA struct {
	msg string
}

type eventB struct {
	msg string
}
