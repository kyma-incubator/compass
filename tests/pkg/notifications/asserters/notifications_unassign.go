package asserters

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/machinebox/graphql"

	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type UnassignNotificationsAsserter struct {
	op                                 string
	state                              *string
	expectedNotificationsCountForOp    int
	useItemsStruct                     bool
	targetObjectID                     string
	sourceObjectID                     string
	localTenantID                      string
	appNamespace                       string
	region                             string
	tenant                             string
	tenantParentCustomer               string
	config                             string
	formationName                      string // used when the test operates with formation different from the one provided in pre  setup
	externalServicesMockMtlsSecuredURL string
	certSecuredGraphQLClient           *graphql.Client
	client                             *http.Client
}

func NewUnassignNotificationsAsserter(expectedNotificationsCountForOp int, targetObjectID string, sourceObjectID string, localTenantID string, appNamespace string, region string, tenant string, tenantParentCustomer string, config string, externalServicesMockMtlsSecuredURL string, client *http.Client) *UnassignNotificationsAsserter {
	return &UnassignNotificationsAsserter{
		op:                                 unassignOperation,
		expectedNotificationsCountForOp:    expectedNotificationsCountForOp,
		targetObjectID:                     targetObjectID,
		sourceObjectID:                     sourceObjectID,
		localTenantID:                      localTenantID,
		appNamespace:                       appNamespace,
		region:                             region,
		tenant:                             tenant,
		tenantParentCustomer:               tenantParentCustomer,
		config:                             config,
		externalServicesMockMtlsSecuredURL: externalServicesMockMtlsSecuredURL,
		client:                             client,
	}
}

func (a *UnassignNotificationsAsserter) WithState(state string) *UnassignNotificationsAsserter {
	a.state = &state
	return a
}

func (a *UnassignNotificationsAsserter) WithUseItemsStruct(useItemsStruct bool) *UnassignNotificationsAsserter {
	a.useItemsStruct = useItemsStruct
	return a
}

func (a *UnassignNotificationsAsserter) WithFormationName(formationName string) *UnassignNotificationsAsserter {
	a.formationName = formationName
	return a
}

func (a *UnassignNotificationsAsserter) WithGQLClient(gqlClient *graphql.Client) *UnassignNotificationsAsserter {
	a.certSecuredGraphQLClient = gqlClient
	return a
}

func (a *UnassignNotificationsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	var formationID string
	if a.formationName != "" {
		formation := fixtures.GetFormationByName(t, ctx, a.certSecuredGraphQLClient, a.formationName, a.tenant)
		formationID = formation.ID
	} else {
		formationID = ctx.Value(context_keys.FormationIDKey).(string)
	}

	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)
	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	notificationsFoundCount := 0
	for _, notification := range notificationsForTarget.Array() {
		op := notification.Get("Operation").String()
		if op == a.op {
			notificationsFoundCount++
			if a.useItemsStruct {
				assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, notification, unassignOperation, formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.tenant, a.tenantParentCustomer, &a.config)
			} else {
				err := verifyFormationAssignmentNotification(t, notification, unassignOperation, formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.config, a.tenant, a.tenantParentCustomer, false, a.state)
				require.NoError(t, err)
			}
		}
	}
	require.Equal(t, a.expectedNotificationsCountForOp, notificationsFoundCount, "expected %d notifications for target object %s", a.expectedNotificationsCountForOp, a.targetObjectID)
}
