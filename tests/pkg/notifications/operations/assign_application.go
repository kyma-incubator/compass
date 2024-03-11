package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
)

type AssignAppToFormationOperation struct {
	applicationID           string
	tenantID                string
	formationName           string
	formationNameContextKey string
	asserters               []asserters.Asserter
}

func NewAssignAppToFormationOperation(applicationID string, tenantID string) *AssignAppToFormationOperation {
	return &AssignAppToFormationOperation{applicationID: applicationID, tenantID: tenantID, formationNameContextKey: context_keys.FormationNameKey}
}

func (o *AssignAppToFormationOperation) WithFormationNameContextKey(formationNAmeContextKey string) *AssignAppToFormationOperation {
	o.formationNameContextKey = formationNAmeContextKey
	return o
}

func (o *AssignAppToFormationOperation) WithAsserters(asserters ...asserters.Asserter) *AssignAppToFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AssignAppToFormationOperation) WithFormationName(formationName string) *AssignAppToFormationOperation {
	o.formationName = formationName
	return o
}

func (o *AssignAppToFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	var formationName string
	if o.formationName != "" {
		formationName = o.formationName
	} else {
		formationName = ctx.Value(o.formationNameContextKey).(string)
	}

	fixtures.AssignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AssignAppToFormationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	var formationName string
	if o.formationName != "" {
		formationName = o.formationName
	} else {
		formationName = ctx.Value(o.formationNameContextKey).(string)
	}

	application := fixtures.GetApplication(t, ctx, gqlClient, o.tenantID, o.applicationID)
	for _, webhook := range application.Webhooks {
		fixtures.DeleteWebhook(t, ctx, gqlClient, o.tenantID, webhook.ID)
	}

	applicationTemplate := fixtures.GetApplicationTemplate(t, ctx, gqlClient, o.tenantID, *application.ApplicationTemplateID)
	for _, webhook := range applicationTemplate.Webhooks {
		fixtures.DeleteWebhook(t, ctx, gqlClient, o.tenantID, webhook.ID)
	}

	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
}

func (o *AssignAppToFormationOperation) Operation() Operation {
	return o
}
