package apptemplate

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

// UnmarshalTenantMappingConfig unmarshalls a string into map[string]interface{}
func UnmarshalTenantMappingConfig(tenantMappingConfigPath string) (map[string]interface{}, error) {
	fileContent, err := os.ReadFile(tenantMappingConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading tenant mapping config file %q", tenantMappingConfigPath)
	}

	var tenantMappingConfig map[string]interface{}
	if err := json.Unmarshal(fileContent, &tenantMappingConfig); err != nil {
		return nil, err
	}

	return tenantMappingConfig, nil
}
