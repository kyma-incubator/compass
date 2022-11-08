package notifications

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Entity struct {
	ID                  string `db:"id" json:"id"`
	TenantID            string `db:"tenant_id" json:"tenant_id"`
	FormationTemplateID string `db:"formation_template_id" json:"formation_template_id"`
	Name                string `db:"name" json:"name"`
}

type FormationNotificationHandler struct {
	Transact              persistence.Transactioner
	DirectorGraphQLClient *gcli.Client
}

func (l *FormationNotificationHandler) HandleCreate(ctx context.Context, data []byte) error {
	entity := Entity{}
	if err := json.Unmarshal(data, &entity); err != nil {
		return errors.Errorf("could not unmarshal app: %s", err)
	}

	log.C(ctx).Infof("Successfully handled create event for formation %v", entity)
	return nil
}

func (l *FormationNotificationHandler) HandleUpdate(ctx context.Context, data []byte) error {
	return nil
}

func (l *FormationNotificationHandler) HandleDelete(ctx context.Context, data []byte) error {
	return nil
}
