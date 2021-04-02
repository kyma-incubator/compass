package auditlog

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
)

type Worker struct {
	svc             proxy.AuditlogService
	client          Client
	auditlogChannel chan proxy.AuditlogMessage
	done            chan bool
	collector       MetricCollector
}

func NewWorker(svc proxy.AuditlogService, auditlogChannel chan proxy.AuditlogMessage, collector MetricCollector) *Worker {
	return &Worker{
		svc:             svc,
		auditlogChannel: auditlogChannel,
		collector:       collector,
	}
}

func (w *Worker) Start(ctx context.Context) {
	logger := log.C(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Infoln("Worker for auditlog message processing has finished")
			return
		case msg := <-w.auditlogChannel:
			logger.Debugf("Read from auditlog channel (size=%d, cap=%d)", len(w.auditlogChannel), cap(w.auditlogChannel))
			w.collector.SetChannelSize(len(w.auditlogChannel))
			ctx := context.WithValue(ctx, correlation.HeadersContextKey, msg.CorrelationIDHeaders)
			err := w.svc.Log(ctx, msg)
			if err != nil {
				logger.WithError(err).Errorf("while saving auditlog message: %v", err)
			}
		}
	}
}
