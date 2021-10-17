package tenantfetcher

import "net/http"

func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (s *GlobalAccountService) SetRetryAttempts(n uint) {
	s.retryAttempts = n
}

func (s *SubaccountService) SetRetryAttempts(n uint) {
	s.retryAttempts = n
}
