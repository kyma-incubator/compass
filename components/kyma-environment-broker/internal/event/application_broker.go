package event

import (
	"context"
	"reflect"
	"sync"
)

type Handler = func(ctx context.Context, ev interface{}) error

type Publisher interface {
	Publish(ctx context.Context, event interface{})
}

// ApplicationEventBroker implements a simple event broker which allows to send event across the application.
type ApplicationEventBroker struct {
	mu sync.Mutex

	handlers map[reflect.Type][]Handler
}

func NewApplicationEventBroker() *ApplicationEventBroker {
	return &ApplicationEventBroker{
		handlers: make(map[reflect.Type][]Handler),
	}
}

func (b *ApplicationEventBroker) Publish(ctx context.Context, ev interface{}) {
	tt := reflect.TypeOf(ev)
	hList, found := b.handlers[tt]
	if found {
		for _, h := range hList {
			h(ctx, ev)
		}
	}
}

func (b *ApplicationEventBroker) Subscribe(evType interface{}, evHandler Handler) {
	tt := reflect.TypeOf(evType)
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, found := b.handlers[tt]; !found {
		b.handlers[tt] = []Handler{}
	}

	b.handlers[tt] = append(b.handlers[tt], evHandler)
}
