package customerrors

import (
	errs "errors"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName(t *testing.T) {
	notFoundErr := NewNotFoundErr("type", "id")
	assert.True(t, errs.Is(notFoundErr, NotFoundErr))

	//_, err := os.Open("/moj/tajny/plik")
	//if err != nil {
	//	// The *os.PathError returned by os.Open is an internal detail.
	//	// To avoid exposing it to the caller, repackage it as a new
	//	// error with the same text. We use the %v formatting verb, since
	//	// %w would permit the caller to unwrap the original *os.PathError.
	//	packedErr := fmt.Errorf("%v", err)
	//	fmt.Printf("oryginal: %+v, repacked: %+v", err, packedErr)
	//}

	err := errors.New("moj error")
	wrappedErr := errors.Wrap(err, "")
	assert.Equal(t, err, errs.Unwrap(errs.Unwrap(wrappedErr)))

	wrappedErr = errors.Wrap(notFoundErr, "")
	assert.True(t, errs.Is(wrappedErr, NotFoundErr))
	fmt.Printf("Before: %+v\n", NotFoundErr)
	assert.True(t, errs.As(wrappedErr, &NotFoundErr))
	fmt.Printf("After: %+v", NotFoundErr)
	//errs.As(notFoundErr, )

}

func TestErrorBuilder_Build(t *testing.T) {
	err := NewErrorBuilder(NotFound).With("type", "application").With("id", "1234").Build()

	fmt.Println(err.Error())
}
