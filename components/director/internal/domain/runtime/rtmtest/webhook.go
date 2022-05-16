package rtmtest

import "github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"

func UnusedWebhookService() func() *automock.WebhookService {
	return func() *automock.WebhookService {
		return &automock.WebhookService{}
	}
}
