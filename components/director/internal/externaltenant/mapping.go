package externaltenant

import (
	"encoding/json"
	"io/ioutil"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type TenantMappingInput struct { //TODO REMOVE
	Name             string `json:"name"`
	ExternalTenantID string `json:"id"`
	Provider         string
}

type MappingOverrides struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Model string `json:"model"`
}

type tenantMap map[string]string

func MapTenants(srcPath, provider string, mappingOverrides MappingOverrides) ([]TenantMappingInput, error) {
	bytes, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading external tenants file")
	}

	var tenantMapSlice []tenantMap
	if err := json.Unmarshal(bytes, &tenantMapSlice); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling external tenants")
	}

	var tenants []TenantMappingInput
	for _, tenantObj := range tenantMapSlice {
		newTenant := map[string]string{
			"ExternalTenantID": tenantObj[mappingOverrides.Id],
			"Name":             tenantObj[mappingOverrides.Name],
			"Provider":         provider,
		}

		tnt := TenantMappingInput{}

		err = mapstructure.Decode(newTenant, &tnt)
		tenants = append(tenants, tnt)
	}

	return tenants, nil
}

func ParseMappingOverrides(srcPath string) (*MappingOverrides, error) {
	bytes, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading external tenants file")
	}

	mappingOverrides := MappingOverrides{}
	err = json.Unmarshal(bytes, &mappingOverrides)
	if err != nil {
		return nil, err
	}
	return &mappingOverrides, nil
}
