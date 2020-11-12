package graphql

type RuntimeContext struct {
	ID    string `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Extended types used by external API

type RuntimeContextPageExt struct {
	RuntimeContextPage
	Data []*RuntimeContextExt `json:"data"`
}

type RuntimeContextExt struct {
	RuntimeContext
	Labels Labels `json:"labels"`
}
