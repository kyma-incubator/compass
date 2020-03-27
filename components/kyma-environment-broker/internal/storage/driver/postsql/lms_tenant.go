package postsql

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"
)

type lmsTenants struct {
	dbsession.Factory
}

func NewLMSTenants(sess dbsession.Factory) *lmsTenants {
	return &lmsTenants{
		Factory: sess,
	}
}

func (s *lmsTenants) FindTenantByName(name, region string) (internal.LMSTenant, bool, error) {
	sess := s.NewReadSession()
	dto, err := sess.GetLMSTenant(name, region)

	switch {
	case err == nil:
		return internal.LMSTenant{
			CreatedAt: dto.CreatedAt,
			Name:      dto.Name,
			Region:    dto.Region,
			ID:        dto.ID,
		}, true, nil
	case err.Code() == dberr.CodeNotFound:
		return internal.LMSTenant{}, false, nil
	default:
		return internal.LMSTenant{}, false, err
	}
}

func (s *lmsTenants) InsertTenant(tenant internal.LMSTenant) error {
	sess := s.NewWriteSession()
	return sess.InsertLMSTenant(dbsession.LMSTenantDTO{
		Name:      tenant.Name,
		Region:    tenant.Region,
		CreatedAt: tenant.CreatedAt,
		ID:        tenant.ID,
	})
}
