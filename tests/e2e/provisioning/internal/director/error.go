package director

type TemporaryError struct {
	message string
}

func (te TemporaryError) Error() string { return te.message }
func (TemporaryError) Temporary() bool  { return true }

func IsTemporaryError(err error) bool {
	nfe, ok := err.(interface {
		Temporary() bool
	})
	return ok && nfe.Temporary()
}
