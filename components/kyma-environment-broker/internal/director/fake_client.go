package director

import (
	"sync"
)

type fakeClient struct {
	mu   sync.Mutex
	URLs map[string]string
}

func NewFakeDirectorClient() *fakeClient {
	return &fakeClient{
		URLs: make(map[string]string, 0),
	}
}

func (dc *fakeClient) SetConsoleURL(runtimeID, URL string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.URLs[runtimeID] = URL
}

// Director Client methods

func (dc *fakeClient) GetConsoleURL(accountID, runtimeID string) (string, error) {
	//for id, URL := range dc.URLs {
	//	if runtimeID == id {
	//		return URL, nil
	//	}
	//}

	return "https://console.e2e-provisioning.gophers.kyma.pro", nil
}
