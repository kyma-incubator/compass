package director

type fakeClient struct{}

func NewFakeDirectorClient() *fakeClient {
	return &fakeClient{}
}

func (dc *fakeClient) GetConsoleURL(accountID, runtimeID string) (string, error) {
	return "", nil
}
