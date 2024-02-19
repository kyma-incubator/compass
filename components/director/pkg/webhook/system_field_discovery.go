package webhook

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type Subscription struct {
	AppURL string `json:"url"`
}

type SubscriptionsResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

func ExecuteSystemFieldDiscoveryWebhook(ctx context.Context, client *http.Client, webhook *model.Webhook) ([]byte, error) {
	webhookURL := webhook.URL
	if webhookURL == nil {
		return nil, errors.Errorf("URL is missing for webhook with id %q", webhook.ID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *webhookURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "error while creating request for webhook with id %q", webhook.ID)
	}

	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "error while executing request for webhook with id %q and URL %q", webhook.ID, *webhookURL)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.C(ctx).Error(err, "Failed to close HTTP response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: expected: %d, but got: %d", http.StatusOK, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTTP response body")
	}

	return respBody, nil
}
