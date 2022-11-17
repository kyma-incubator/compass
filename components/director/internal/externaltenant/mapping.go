package externaltenant

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/pkg/errors"
)

// MapTenants missing godoc
func MapTenants(tenantsDirectoryPath, defaultTenantRegion string) ([]model.BusinessTenantMappingInput, error) {
	files, err := os.ReadDir(tenantsDirectoryPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading directory with tenant files [%s]", tenantsDirectoryPath)
	}

	var outputTenants []model.BusinessTenantMappingInput
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			return nil, apperrors.NewInternalError(fmt.Sprintf("unsupported file format [%s]", filepath.Ext(f.Name())))
		}

		bytes, err := os.ReadFile(tenantsDirectoryPath + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "while reading tenants file [%s]", tenantsDirectoryPath+f.Name())
		}

		var tenantsFromFile []model.BusinessTenantMappingInput
		if err := json.Unmarshal(bytes, &tenantsFromFile); err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling tenants from file [%s]", tenantsDirectoryPath+f.Name())
		}

		for i := range tenantsFromFile {
			tenantsFromFile[i].Provider = f.Name()
			if tenantsFromFile[i].Region == "" && tenantsFromFile[i].Type == string(tenant.Subaccount) {
				tenantsFromFile[i].Region = defaultTenantRegion
			}
		}

		outputTenants = append(outputTenants, tenantsFromFile...)
	}

	return outputTenants, nil
}
