package auditlog

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
)

type Auditlog interface {
	Log(request, response string, claims proxy.Claims) error
}

type AuditLogWorker struct {
	svc             Auditlog
	client          Client
	auditlogChannel chan AuditlogMessage
	done            chan bool
}

func NewWorker(svc Auditlog, auditlogChannel chan AuditlogMessage, done chan bool) *AuditLogWorker {
	return &AuditLogWorker{
		svc:             svc,
		auditlogChannel: auditlogChannel,
		done:            done,
	}
}

func (w *AuditLogWorker) Start() {
	for {
		select {
		case <-w.done:
			return
		case log := <-w.auditlogChannel:
			err := w.svc.Log(log.Request, log.Response, log.Claims)
			if err != nil {
				fmt.Printf("error while saving auditlog: %s\n", err.Error())
			}
		}
	}
}
