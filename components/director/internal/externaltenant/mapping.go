package externaltenant

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/pkg/errors"
)

func MapTenants(srcPath, provider string) ([]model.BusinessTenantMappingInput, error) {
	bytes, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading external tenants file")
	}

	var tenants []model.BusinessTenantMappingInput
	if err := json.Unmarshal(bytes, &tenants); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling external tenants from file %s", srcPath)
	}

	for i := range tenants {
		tenants[i].Provider = provider
	}

	return tenants, nil
}
