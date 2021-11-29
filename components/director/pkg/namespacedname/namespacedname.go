package namespacedname

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/types"
)

const defaultNamespace = "default"

// Parse is responsible to convert string into NamespacedName structure,
// comprising the resource name and mandatory namespace
func Parse(value string) (types.NamespacedName, error) {
	if value == "" {
		return types.NamespacedName{}, errors.New("the value cannot be empty")
	}

	parts := strings.Split(value, string(types.Separator))

	if len(parts) > 2 {
		return types.NamespacedName{}, errors.New(fmt.Sprintf("the given value: %s is not in the expected format - <namespace>/<name>", value))
	}

	// If a single value is provided we assume that's resource name and the namespace is default
	if len(parts) == 1 && parts[0] != "" {
		return types.NamespacedName{
			Name:      parts[0],
			Namespace: defaultNamespace,
		}, nil
	}

	if len(parts) == 2 && parts[1] == "" {
		return types.NamespacedName{}, errors.New("resource name should not be empty")
	}

	namespace := parts[0]
	if namespace == "" {
		namespace = defaultNamespace
	}

	return types.NamespacedName{
		Namespace: namespace,
		Name:      parts[1],
	}, nil
}
