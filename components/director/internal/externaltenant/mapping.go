package externaltenant

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

type TenantMappingInput struct {
	Name             string `json:"name"`
	ExternalTenantID string `json:"id"`
	Provider         string
}

func MapTenants(srcPath, provider string) ([]TenantMappingInput, error) {
	bytes, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading external tenants file")
	}

	var tenants []TenantMappingInput
	if err := json.Unmarshal(bytes, &tenants); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling external tenants from file %s", srcPath)
	}

	for i := range tenants {
		tenants[i].Provider = provider
	}

	return tenants, nil
}
