package notifications

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type FAEntity struct {
	ID          string          `db:"id" json:"id"`
	FormationID string          `db:"formation_id" json:"formation_id"`
	TenantID    string          `db:"tenant_id" json:"tenant_id"`
	Source      string          `db:"source" json:"source"`
	SourceType  string          `db:"source_type" json:"source_type"`
	Target      string          `db:"target" json:"target"`
	TargetType  string          `db:"target_type" json:"target_type"`
	State       string          `db:"state" json:"state"`
	Value       json.RawMessage `db:"value" json:"value"`
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
	fa := FAEntity{}
	if err := json.Unmarshal(data, &fa); err != nil {
		return errors.Errorf("could not unmarshal app: %s", err)
	}

	logger := log.C(ctx).WithFields(logrus.Fields{
		"formationID": fa.ID,
	})
	ctx = log.ContextWithLogger(ctx, logger)

	tx, err := l.Transact.Begin()
	if err != nil {
		log.C(ctx).Errorf("Error while opening transaction in formation_assignment_handler when updating FA with ID: %q and error: %s", fa.ID, err)
		return err
	}
	defer l.Transact.RollbackUnlessCommitted(ctx, tx)

	// 1. gjson get di-tenant-details -> if not exists return
	// 2. get di-tenant-details.tenant-id && di-tenant-details.mngmnt-url
	// 3. update App with id = FA.Target and set the two fields from 2.
	// 4. sjson remove di-tenant-details from config -> if config is empty object -> set to null else update with result from sjson

	diTenantDetails := gjson.GetBytes(fa.Value, "di-tenant-details")
	if !diTenantDetails.Exists() {
		return nil
	}
	if !diTenantDetails.IsObject() {
		return errors.Errorf("'di-tenant-details' from config must be an object")
	}

	tenantID := diTenantDetails.Get("tenant-id")
	if !tenantID.Exists() {
		return errors.Errorf("'di-tenant-details.tenant-id' does not exists")
	}

	managementURL := diTenantDetails.Get("management-url")
	if !managementURL.Exists() {
		return errors.Errorf("'di-tenant-details.management-url' does not exists")
	}

	log.C(ctx).Infof("Updating fields of Application with ID: %q with localTenantID: %q and baseURL: %q", fa.Target, tenantID.String(), managementURL.String())
	_, err = tx.ExecContext(ctx, "UPDATE applications SET base_url = $1, local_tenant_id = $2 WHERE id = $3",
		managementURL.String(), tenantID.String(), fa.Target)
	if err != nil {
		return errors.Wrapf(err, "while updating Application with ID: %q", fa.Target)
	}

	newValue, err := sjson.DeleteBytes(fa.Value, "di-tenant-details")
	if err != nil {
		return errors.Wrap(err, "while deleting with sjson")
	}

	if len(gjson.ParseBytes(newValue).Map()) == 0 {
		log.C(ctx).Infof("Updating value of FA with ID: %q with null config", fa.ID)
		_, err = tx.ExecContext(ctx, "UPDATE formation_assignments SET value = NULL WHERE id = $1", fa.ID)
		if err != nil {
			return errors.Wrapf(err, "while updating value of FA with ID: %q", fa.ID)
		}
	} else {
		log.C(ctx).Infof("Updating value of FA with ID: %q with new value: %s", fa.ID, string(newValue))
		_, err = tx.ExecContext(ctx, "UPDATE formation_assignments SET value = $1 WHERE id = $2", string(newValue), fa.ID)
		if err != nil {
			return errors.Wrapf(err, "while updating new value of FA with ID: %q", fa.ID)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.C(ctx).Errorf("Error while committing transaction in formation_assignment_handler when updating FA with ID: %q and error: %s", fa.ID, err)
		return err
	}

	log.C(ctx).Infof("Successfully handled update event for formation assignment %v", fa)
	return nil
}

func (l *FANotificationHandler) HandleDelete(ctx context.Context, data []byte) error {
	return nil
}
