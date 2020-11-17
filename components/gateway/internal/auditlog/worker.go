package auditlog

import (
	"context"
	"log"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
)

type Worker struct {
	svc             proxy.AuditlogService
	client          Client
	auditlogChannel chan proxy.AuditlogMessage
	done            chan bool
}

func NewWorker(svc proxy.AuditlogService, auditlogChannel chan proxy.AuditlogMessage, done chan bool) *Worker {
	return &Worker{
		svc:             svc,
		auditlogChannel: auditlogChannel,
		done:            done,
	}
}

func (w *Worker) Start() {
	ctx := context.Background()
	for {
		select {
		case <-w.done:
			log.Println("Worker for auditlog message processing has finished")
			ctx.Done()
			return
		case msg := <-w.auditlogChannel:
			log.Printf("Read from auditlog channel (size=%d, cap=%d)", len(w.auditlogChannel), cap(w.auditlogChannel))
			ctx := context.WithValue(ctx, correlation.HeadersContextKey, msg.CorrelationIDHeaders)
			err := w.svc.Log(ctx, msg)
			if err != nil {
				log.Printf("Error while saving auditlog message with error: %s", err.Error())
			}
		}
	}
}
