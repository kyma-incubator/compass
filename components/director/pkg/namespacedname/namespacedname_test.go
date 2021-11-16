package namespacedname

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {

	t.Run("should parse string", func(t *testing.T) {
		// given
		namespace := "namespace"
		name := "name"
		str := fmt.Sprintf("%s/%s", namespace, name)

		// when
		namespacedname := Parse(str)

		// then
		assert.Equal(t, namespacedname.Namespace, namespace)
		assert.Equal(t, namespacedname.Name, name)
	})

	t.Run("should return empty namespacedname", func(t *testing.T) {
		// given
		namespace := ""
		name := ""
		str := fmt.Sprintf("%s/%s", namespace, name)

		// when
		namespacedname := Parse(str)

		// then
		assert.Equal(t, namespacedname.Namespace, "default")
		assert.Equal(t, namespacedname.Name, name)
	})

	t.Run("should return name with default namespace", func(t *testing.T) {
		// given
		name := "name"
		str := fmt.Sprintf("%s", name)

		// when
		namespacedname := Parse(str)

		// then
		assert.Equal(t, namespacedname.Namespace, "default")
		assert.Equal(t, namespacedname.Name, name)
	})

	t.Run("should return empty name if slash provided", func(t *testing.T) {
		// given
		str := fmt.Sprintf("/")

		// when
		namespacedname := Parse(str)

		// then
		assert.Equal(t, namespacedname.Namespace, "default")
		assert.Equal(t, namespacedname.Name, "")
	})

	t.Run("should return empty name if empty string provided", func(t *testing.T) {
		// given
		str := fmt.Sprintf("")

		// when
		namespacedname := Parse(str)

		// then
		assert.Equal(t, namespacedname.Namespace, "default")
		assert.Equal(t, namespacedname.Name, "")
	})
}
