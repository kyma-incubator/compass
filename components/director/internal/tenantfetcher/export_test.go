package tenantfetcher

import "net/http"

func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (s *Service) SetRetryAttempts(n uint) {
	s.retryAttempts = n
}
