package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNullableBool(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		input := true
		// WHEN
		result := NewNullableBool(&input)
		// THEN
		assert.True(t, result.Valid)
		assert.Equal(t, input, result.Bool)
	})

	t.Run("return not valid when nil bool", func(t *testing.T) {
		// WHEN
		result := NewNullableBool(nil)
		// THEN
		assert.False(t, result.Valid)
	})
}

func TestNewNullableString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		text := "lorem ipsum"
		// WHEN
		result := NewNullableString(&text)
		// THEN
		assert.True(t, result.Valid)
		assert.Equal(t, text, result.String)
	})

	t.Run("success when empty string", func(t *testing.T) {
		// GIVEN
		text := ""
		// WHEN
		result := NewNullableString(&text)
		// THEN
		assert.False(t, result.Valid)
		assert.Equal(t, text, result.String)
	})

	t.Run("return not valid when nil string", func(t *testing.T) {
		// WHEN
		result := NewNullableString(nil)
		// THEN
		assert.False(t, result.Valid)
	})
}
