package auditlog

import (
	"context"
	"log"

	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
)

type AuditlogService interface {
	Log(ctx context.Context, request, response string, claims proxy.Claims) error
}

type Worker struct {
	svc             AuditlogService
	client          Client
	auditlogChannel chan Message
	done            chan bool
}

func NewWorker(svc AuditlogService, auditlogChannel chan Message, done chan bool) *Worker {
	return &Worker{
		svc:             svc,
		auditlogChannel: auditlogChannel,
		done:            done,
	}
}

func (w *Worker) Start() {
	for {
		select {
		case <-w.done:
			log.Println("Worker for auditlog message processing has finished")
			return
		case msg := <-w.auditlogChannel:
			log.Printf("Read from auditlog channel (size=%d, cap=%d)", len(w.auditlogChannel), cap(w.auditlogChannel))
			err := w.svc.Log(msg.Context, msg.Request, msg.Response, msg.Claims)
			if err != nil {
				log.Printf("Error while saving auditlog message with error: %s", err.Error())
			}
		}
	}
}
