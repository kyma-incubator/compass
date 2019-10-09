package oauth20

func (s *service) SetHTTPClient(httpCli HTTPClient) {
	s.httpCli = httpCli
}
