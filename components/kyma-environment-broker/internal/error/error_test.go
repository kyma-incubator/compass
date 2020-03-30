package error

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWrapf(t *testing.T) {
	// given
	err1 := fmt.Errorf("some error: %s", "argErr")
	err2 := fmt.Errorf("some error: %s", "argErr")
	err3 := NewTemporaryError("some error: %s", fmt.Errorf("argErr"))

	// when
	e1 := errors.Wrapf(err1, "wrap err %s", "arg1")
	e2 := AsTemporaryError(err2, "wrap err %s", "arg1")
	e3 := errors.Wrapf(err3, "wrap err %s", "arg1")

	// then
	assert.False(t, IsTemporaryError(e1))
	assert.True(t, IsTemporaryError(e2))
	assert.True(t, IsTemporaryError(e3))

	assert.Equal(t, "wrap err arg1: some error: argErr", e1.Error())
	assert.Equal(t, "wrap err arg1: some error: argErr", e2.Error())
	assert.Equal(t, "wrap err arg1: some error: argErr", e3.Error())
}
