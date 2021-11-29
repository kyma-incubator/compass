package namespacedname

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	namespace := "namespace"
	name := "resource-name"
	emptyStr := ""

	t.Run("Should return error if empty string is provided", func(t *testing.T) {
		// given
		str := emptyStr

		// when
		namespacedname, err := Parse(str)

		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "the value cannot be empty")
		require.Empty(t, namespacedname)
	})

	t.Run("Should return error when the string is not in the expected format", func(t *testing.T) {
		// given
		str := fmt.Sprintf("%s/%s/wrong", namespace, name)

		// when
		namespacedname, err := Parse(str)

		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("the given value: %s is not in the expected format - <namespace>/<name>", str))
		require.Empty(t, namespacedname)
	})

	t.Run("If single value is provided should return it as resource name and use default namespace", func(t *testing.T) {
		// given
		str := name

		// when
		namespacedname, err := Parse(str)

		// then
		require.NoError(t, err)
		require.Equal(t, namespacedname.Namespace, "default")
		require.Equal(t, namespacedname.Name, name)
	})

	t.Run("Should return error when only slash is provided", func(t *testing.T) {
		// given
		str := "/"

		// when
		namespacedname, err := Parse(str)

		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "resource name should not be empty")
		require.Empty(t, namespacedname)
	})

	t.Run("If namespace is empty use the provided resource name with default namespace", func(t *testing.T) {
		// given
		str := fmt.Sprintf("%s/%s", emptyStr, name)

		// when
		namespacedname, err := Parse(str)

		// then
		require.NoError(t, err)
		require.Equal(t, namespacedname.Namespace, "default")
		require.Equal(t, namespacedname.Name, name)
	})

	t.Run("Successful when both namespace and resource name are provided", func(t *testing.T) {
		// given
		str := fmt.Sprintf("%s/%s", namespace, name)

		// when
		namespacedname, err := Parse(str)

		// then
		require.NoError(t, err)
		require.Equal(t, namespacedname.Namespace, namespace)
		require.Equal(t, namespacedname.Name, name)
	})
}
