package connector

// Mock implementation of EventServiceBaseURLProvider
type eventBaseURLProvider struct {
	eventBaseUrl string
}

func (e eventBaseURLProvider) EventServiceBaseURL() (string, error) {
	// TODO: call Director for getting events base url
	return e.eventBaseUrl, nil
}
