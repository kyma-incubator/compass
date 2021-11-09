package model

type NSError struct {
	message string
}

func NewNSError(msg string) *NSError {
	return &NSError{
		message: msg,
	}
}

func (m *NSError) Error() string {
	return m.message
}