package model

// Identifiable is an interface that can be used to identify an object.
type Identifiable interface {
	GetID() string
}
