package notifications

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type FAEntity struct {
	ID          string `db:"id" json:"id"`
	FormationID string `db:"formation_id" json:"formation_id"`
	TenantID    string `db:"tenant_id" json:"tenant_id"`
	Source      string `db:"source" json:"source"`
	SourceType  string `db:"source_type" json:"source_type"`
	Target      string `db:"target" json:"target"`
	TargetType  string `db:"target_type" json:"target_type"`
	State       string `db:"state" json:"state"`
	Value       string `db:"value" json:"value"`
}
type FANotificationHandler struct {
	Transact                         persistence.Transactioner
	DirectorGraphQLClient            *gcli.Client
	DirectorCertSecuredGraphQLClient *gcli.Client
}

func (l *FANotificationHandler) HandleCreate(ctx context.Context, data []byte) error {
	return nil
}

func (l *FANotificationHandler) HandleUpdate(ctx context.Context, data []byte) error {
	entity := FAEntity{}
	if err := json.Unmarshal(data, &entity); err != nil {
		return errors.Errorf("could not unmarshal app: %s", err)
	}

	tx, err := l.Transact.Begin()
	if err != nil {
		log.C(ctx).Errorf("Error while opening transaction in formation_assignment_handler when updating FA with ID: %q and error: %s", entity.ID, err)
		return err
	}
	defer l.Transact.RollbackUnlessCommitted(ctx, tx)

	err = tx.Commit()
	if err != nil {
		log.C(ctx).Errorf("Error while committing transaction in formation_assignment_handler when updating FA with ID: %q and error: %s", entity.ID, err)
		return err
	}

	log.C(ctx).Infof("Successfully handled update event for formation assignment %v", entity)
	return nil
}

func (l *FANotificationHandler) HandleDelete(ctx context.Context, data []byte) error {
	return nil
}
