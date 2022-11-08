package notifications

import (
	"context"
	"github.com/lib/pq"
)

type NotificationHandler interface {
	HandleCreate(ctx context.Context, data []byte) error
	HandleUpdate(ctx context.Context, data []byte) error
	HandleDelete(ctx context.Context, data []byte) error
}

type NotificationListener interface {
	Listen(channel string) error
	Ping() error
	Close() error
	NotificationChannel() <-chan *pq.Notification
}
