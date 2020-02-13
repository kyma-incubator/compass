package runtime

func (r *ComponentsListProvider) WithHTTPClient(doer HTTPDoer) *ComponentsListProvider {
	r.httpClient = doer

	return r
}
