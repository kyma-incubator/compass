package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlizer "github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Entity struct {
	ID                  string `db:"id" json:"id"`
	TenantID            string `db:"tenant_id" json:"tenant_id"`
	FormationTemplateID string `db:"formation_template_id" json:"formation_template_id"`
	Name                string `db:"name" json:"name"`
}

const (
	DIFormationTemplateID     = "686d42be-d944-4b63-be72-047603df06e6"
	DIApplicationTemplateName = "SAP Data Ingestion"
)

type FormationNotificationHandler struct {
	Transact                         persistence.Transactioner
	DirectorGraphQLClient            *gcli.Client
	DirectorCertSecuredGraphQLClient *gcli.Client
	Graphqlizer                      gqlizer.Graphqlizer
	GQLFieldsProvider                gqlizer.GqlFieldsProvider
}

func (l *FormationNotificationHandler) HandleCreate(ctx context.Context, data []byte) error {
	formation := Entity{}
	if err := json.Unmarshal(data, &formation); err != nil {
		return errors.Errorf("could not unmarshal app: %s", err)
	}

	logger := log.C(ctx).WithFields(logrus.Fields{
		"formationID": formation.ID,
	})
	ctx = log.ContextWithLogger(ctx, logger)

	if formation.FormationTemplateID != DIFormationTemplateID {
		log.C(ctx).Infof("Formation %v is not DI formation. Nothing to process.", formation)
		return nil
	}

	tx, err := l.Transact.Begin()
	if err != nil {
		log.C(ctx).Errorf("Error while opening transaction in formation_notification_handler when creating formation with ID: %q and error: %s", formation.ID, err)
		return err
	}
	defer l.Transact.RollbackUnlessCommitted(ctx, tx)

	log.C(ctx).Infof("Getting external tenant ID for formation with ID: %q and internal tenant ID: %q", formation.ID, formation.TenantID)
	var externalTenantID string
	err = tx.GetContext(ctx, &externalTenantID, "SELECT external_tenant from business_tenant_mappings WHERE id = $1", formation.TenantID)
	if err != nil {
		return errors.Wrapf(err, "while getting external tenant ID for Formation with ID: %q and internal tenant ID: %q", formation.ID, formation.TenantID)
	}

	if externalTenantID == "" {
		return errors.Errorf("external tenant ID for Formation with ID: %q and internal tenant ID: %q can not be empty", formation.ID, formation.TenantID)
	}

	err = tx.Commit()
	if err != nil {
		log.C(ctx).Errorf("Error while committing transaction in formation_notification_handler when creating formation with ID: %q and error: %s", formation.ID, err)
		return err
	}

	appFromTmplSrc := graphql.ApplicationFromTemplateInput{
		TemplateName: DIApplicationTemplateName, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       formation.Name,
			},
			{
				Placeholder: "display-name",
				Value:       formation.Name,
			},
		},
	}

	log.C(ctx).Info("Start registering Application from Application Template")

	appFromTmplSrcGQL, err := l.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	if err != nil {
		return errors.Wrap(err, "while generating graphql payload")
	}

	createAppFromTmplFirstRequest := l.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	outputSrcApp := graphql.ApplicationExt{}
	err = RunOperationWithCustomTenant(ctx, l.DirectorCertSecuredGraphQLClient, externalTenantID, createAppFromTmplFirstRequest, &outputSrcApp)
	if err != nil {
		log.C(ctx).Errorf("Error while registering Application: %s", err)
		return err
	}
	log.C(ctx).Info("Successfully registered Application from Application Template. Sleeping for 15s to avoid race condition with BTP Cockpit...")
	time.Sleep(15 * time.Second)

	log.C(ctx).Infof("Assigning Application with ID: %q to formation with ID: %q", outputSrcApp.ID, formation.ID)

	assignReq := l.FixAssignFormationRequest(outputSrcApp.ID, string(graphql.FormationObjectTypeApplication), formation.Name)
	var assignFormation graphql.Formation
	err = RunOperationWithCustomTenant(ctx, l.DirectorCertSecuredGraphQLClient, externalTenantID, assignReq, &assignFormation)
	if err != nil {
		log.C(ctx).Errorf("Error while assigning Application with ID: %q to Formation with ID: %q and error: %s", outputSrcApp.ID, formation.ID, err)
		return err
	}

	log.C(ctx).Infof("Successfully assigned Application with ID: %q to formation with ID: %q", outputSrcApp.ID, formation.ID)

	log.C(ctx).Infof("Successfully handled create event for formation %v", formation)
	return nil
}

func (l *FormationNotificationHandler) HandleUpdate(ctx context.Context, data []byte) error {
	return nil
}

func (l *FormationNotificationHandler) HandleDelete(ctx context.Context, data []byte) error {
	return nil
}

func (l *FormationNotificationHandler) FixRegisterApplicationFromTemplate(applicationFromTemplateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplicationFromTemplate(in: %s) {
					%s
				}
			}`,
			applicationFromTemplateInputInGQL, l.GQLFieldsProvider.ForApplication()))
}

func RunOperationWithCustomTenant(ctx context.Context, cli *gcli.Client, tenant string, req *gcli.Request, resp interface{}) error {
	return NewOperation(ctx).WithTenant(tenant).Run(req, cli, resp)
}

type Operation struct {
	ctx context.Context

	tenant      string
	queryParams map[string]string
}

func (o *Operation) WithTenant(tenant string) *Operation {
	o.tenant = tenant
	return o
}

func (o *Operation) Run(req *gcli.Request, cli *gcli.Client, resp interface{}) error {
	m := resultMapperFor(&resp)
	req.Header.Set("Tenant", o.tenant)

	return withRetryOnTemporaryConnectionProblems(func() error {
		return cli.Run(o.ctx, req, &m)
	})
}

// resultMapperFor returns generic object that can be passed to Run method for storing response.
// In GraphQL, set `result` alias for your query
func resultMapperFor(target interface{}) genericGQLResponse {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		panic("target has to be a pointer")
	}
	return genericGQLResponse{
		Result: target,
	}
}

type genericGQLResponse struct {
	Result interface{} `json:"result"`
}

func withRetryOnTemporaryConnectionProblems(risky func() error) error {
	return retry.Do(risky, retry.Attempts(7), retry.Delay(time.Second), retry.OnRetry(func(n uint, err error) {
		logrus.WithField("component", "TestContext").Warnf("OnRetry: attempts: %d, error: %v", n, err)

	}), retry.LastErrorOnly(true), retry.RetryIf(func(err error) bool {
		return strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "connection reset by peer")
	}))
}

func NewOperation(ctx context.Context) *Operation {
	return &Operation{
		ctx:         ctx,
		queryParams: map[string]string{},
	}
}

func (l *FormationNotificationHandler) FixAssignFormationRequest(objID, objType, formationName string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{
			  result: assignFormation(objectID:"%s",objectType: %s ,formation: {name: "%s"}){
				%s
			  }
			}`, objID, objType, formationName, l.GQLFieldsProvider.ForFormation()))
}
