package deprovisioning

type instanceNotFoundError struct {
	message string
}

func (inf instanceNotFoundError) Error() string {
	return inf.message
}

func newInstanceNotFoundError(message string) *instanceNotFoundError {
	return &instanceNotFoundError{message: message}
}

type instanceGetError struct {
	message string
}

func (inf instanceGetError) Error() string {
	return inf.message
}

func newInstanceGetError(message string) *instanceGetError {
	return &instanceGetError{message: message}
}

