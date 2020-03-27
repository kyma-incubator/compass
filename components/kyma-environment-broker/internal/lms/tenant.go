package lms

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

type TenantStorage interface {
	FindTenantByName(name, region string) (internal.LMSTenant, bool, error)
	InsertTenant(tenant internal.LMSTenant) error
}

//go:generate mockery -name=TenantCreator -output=automock -outpkg=automock -case=underscore

type TenantCreator interface {
	CreateTenant(input CreateTenantInput) (o CreateTenantOutput, err error)
}

type manager struct {
	tenantStorage TenantStorage
	lmsClient     TenantCreator
	log           logrus.FieldLogger
}

func NewTenantManager(storage TenantStorage, lmsClient TenantCreator, log logrus.FieldLogger) *manager {
	return &manager{
		tenantStorage: storage,
		lmsClient:     lmsClient,
		log:           log,
	}
}

// ProvideLMSTenantID returns existing tenant ID or creates new one (if not exists)
func (c *manager) ProvideLMSTenantID(name, region string) (string, error) {
	tenant, exists, err := c.tenantStorage.FindTenantByName(name, region)
	if err != nil {
		return "", errors.Wrapf(err, "while checking if tenant is already created")
	}

	if !exists {
		output, err := c.lmsClient.CreateTenant(CreateTenantInput{
			Name:   name,
			Region: region, //todo: implement mapping
		})
		if err != nil {
			return "", errors.Wrapf(err, "while creating tenant in lms")
		}

		// it is important to save the tenant ID because tenant creation means creation of a cluster.
		err = wait.PollImmediate(3*time.Second, 30*time.Second, func() (bool, error) {
			err := c.tenantStorage.InsertTenant(internal.LMSTenant{
				ID:        output.ID,
				Name:      name,
				Region:    region,
				CreatedAt: time.Now(),
			})
			if err != nil {
				c.log.Warn(errors.Wrapf(err, "while saving lms tenant %s with ID %s", name, output.ID).Error())
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return "", errors.Wrapf(err, "while saving tenant to storage")
		}
		return output.ID, nil
	}

	return tenant.ID, nil
}
