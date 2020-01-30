package externaltenant

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/pkg/errors"
)

func MapTenants(tenantsDirectoryPath string) ([]model.BusinessTenantMappingInput, error) {

	files, err := ioutil.ReadDir(tenantsDirectoryPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading directory with tenant files [%s]", tenantsDirectoryPath)
	}

	var outputTenants []model.BusinessTenantMappingInput
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			return nil, errors.New(fmt.Sprintf("unsupported file format [%s]", filepath.Ext(f.Name())))
		}

		bytes, err := ioutil.ReadFile(tenantsDirectoryPath + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "while reading tenants file [%s]", tenantsDirectoryPath+f.Name())
		}

		var tenantsFromFile []model.BusinessTenantMappingInput
		if err := json.Unmarshal(bytes, &tenantsFromFile); err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling tenants from file [%s]", tenantsDirectoryPath+f.Name())
		}

		for i := range tenantsFromFile {
			tenantsFromFile[i].Provider = f.Name()
		}

		outputTenants = append(outputTenants, tenantsFromFile...)
	}

	return outputTenants, nil
}
