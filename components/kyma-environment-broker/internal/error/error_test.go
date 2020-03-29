package error

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapf(t *testing.T) {
	// given
	err1 := fmt.Errorf("some error: %s", "argErr")
	err2 := NewTemporaryError("some error", fmt.Errorf("argErr"))
	err3 := NewTemporaryError("some error", fmt.Errorf("argErr"))

	// when
	e1 := Wrapf(err1, "wrap err %s", "arg1")
	e2 := Wrapf(err2, "wrap err %s", "arg1")
	e3 := Wrapf(err3, "wrap err")
	e4 := Wrapf(e3, "another wrap error")

	// then
	assert.False(t, IsTemporaryError(e1))
	assert.True(t, IsTemporaryError(e2))
	assert.True(t, IsTemporaryError(e3))
	assert.True(t, IsTemporaryError(e4))

	assert.Equal(t, "some error: argErr: wrap err arg1", e1.Error())
	assert.Equal(t, "some error: argErr: wrap err arg1", e2.Error())
	assert.Equal(t, "some error: argErr: wrap err", e3.Error())
	assert.Equal(t, "some error: argErr: wrap err: another wrap error", e4.Error())
}
