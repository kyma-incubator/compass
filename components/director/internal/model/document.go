package model

type Document struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Kind  *string `json:"kind"`
	// TODO: Replace with actual model
}

type DocumentInput struct {
	Title       string  `json:"title"`
	DisplayName string  `json:"displayName"`
	Description string  `json:"description"`
	Kind        *string `json:"kind"`
	Data        *[]byte `json:"data"`
	// TODO: Replace with actual model
}

func (d *DocumentInput) ToDocument() *Document {
	// TODO: Replace with actual model
	return &Document{}
}
