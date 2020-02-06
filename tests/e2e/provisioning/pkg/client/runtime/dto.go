package runtime

import schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

type runtimeStatusResponse struct {
	Result schema.RuntimeStatus `json:"result"`
}
