package rtmtest

import "github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"

// UnusedWebhookService returns a mock webhook service that does not expect to get called
func UnusedWebhookService() func() *automock.WebhookService {
	return func() *automock.WebhookService {
		return &automock.WebhookService{}
	}
}
