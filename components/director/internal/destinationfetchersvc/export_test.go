package destinationfetchersvc

import "net/http"

func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}
