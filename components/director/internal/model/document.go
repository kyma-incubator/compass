package model

type DocumentInput struct {
	Title        string             `json:"title"`
	DisplayName  string             `json:"displayName"`
	Description  string             `json:"description"`
	Kind         *string            `json:"kind"`
	Data         *[]byte              `json:"data"`
	// TODO: Replace with actual model
}

