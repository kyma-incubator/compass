package graphql

// RuntimeContext missing godoc
type RuntimeContext struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RuntimeContextPageExt is an extended types used by external API
type RuntimeContextPageExt struct {
	RuntimeContextPage
	Data []*RuntimeContextExt `json:"data"`
}

// RuntimeContextExt missing godoc
type RuntimeContextExt struct {
	RuntimeContext
	Labels Labels `json:"labels"`
}
