package asserters

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/machinebox/graphql"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/sjson"

	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	unassignOperation = "unassign"
	assignOperation   = "assign"
)

type NotificationsAsserter struct {
	expectedNotificationsCount         int
	useItemsStruct                     bool
	assertTrustDetails                 bool
	expectedSubjects                   []string
	op                                 string
	state                              *string
	targetObjectID                     string
	sourceObjectID                     string
	localTenantID                      string
	appNamespace                       string
	region                             string
	tenant                             string
	tenantParentCustomer               string
	config                             *string
	formationName                      string // used when the test operates with formation different from the one provided in pre  setup
	externalServicesMockMtlsSecuredURL string
	certSecuredGraphQLClient           *graphql.Client
	client                             *http.Client
}

func NewNotificationsAsserter(expectedNotificationsCount int, op string, targetObjectID, sourceObjectID string, localTenantID string, appNamespace string, region string, tenant string, tenantParentCustomer string, config *string, externalServicesMockMtlsSecuredURL string, client *http.Client) *NotificationsAsserter {
	return &NotificationsAsserter{
		expectedNotificationsCount:         expectedNotificationsCount,
		op:                                 op,
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

func (a *NotificationsAsserter) WithState(state string) *NotificationsAsserter {
	a.state = &state
	return a
}

func (a *NotificationsAsserter) WithFormationName(formationName string) *NotificationsAsserter {
	a.formationName = formationName
	return a
}

func (a *NotificationsAsserter) WithGQLClient(gqlClient *graphql.Client) *NotificationsAsserter {
	a.certSecuredGraphQLClient = gqlClient
	return a
}

func (a *NotificationsAsserter) WithUseItemsStruct(useItemsStruct bool) *NotificationsAsserter {
	a.useItemsStruct = useItemsStruct
	return a
}

func (a *NotificationsAsserter) WithAssertTrustDetails(assertTrustDetails bool) *NotificationsAsserter {
	a.assertTrustDetails = assertTrustDetails
	return a
}

func (a *NotificationsAsserter) WithExpectedSubjects(expectedSubjects []string) *NotificationsAsserter {
	a.expectedSubjects = expectedSubjects
	return a
}

func (a *NotificationsAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	var formationID string
	if a.formationName != "" {
		formation := fixtures.GetFormationByName(t, ctx, a.certSecuredGraphQLClient, a.formationName, a.tenant)
		formationID = formation.ID
	} else {
		formationID = ctx.Value(context_keys.FormationIDKey).(string)
	}

	body := getNotificationsFromExternalSvcMock(t, a.client, a.externalServicesMockMtlsSecuredURL)
	assertNotificationsCount(t, body, a.targetObjectID, a.expectedNotificationsCount)
	if a.expectedNotificationsCount == 0 {
		return
	}

	notificationsForTarget := gjson.GetBytes(body, a.targetObjectID)
	assignNotificationAboutSource := notificationsForTarget.Array()[0]
	if a.useItemsStruct {
		assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, assignNotificationAboutSource, assignOperation, formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.tenant, a.tenantParentCustomer, a.config)
		if a.assertTrustDetails {
			require.NotEmpty(t, a.expectedSubjects)
			assertTrustDetailsForTargetAndNoTrustDetailsForSource(t, assignNotificationAboutSource, a.expectedSubjects)
		}
	} else {
		err := verifyFormationAssignmentNotification(t, assignNotificationAboutSource, assignOperation, formationID, a.sourceObjectID, a.localTenantID, a.appNamespace, a.region, a.config, a.tenant, a.tenantParentCustomer, false, a.state)
		require.NoError(t, err)
	}
}

func getNotificationsFromExternalSvcMock(t *testing.T, client *http.Client, ExternalServicesMockMtlsSecuredURL string) []byte {
	t.Logf("Getting formation notifications recieved in external services mock")
	resp, err := client.Get(ExternalServicesMockMtlsSecuredURL + "/formation-callback")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusOK, string(body)))
	return body
}

func assertNotificationsCount(t *testing.T, body []byte, objectID string, count int) {
	notifications := gjson.GetBytes(body, objectID)
	if count > 0 {
		require.True(t, notifications.Exists())
		require.Len(t, notifications.Array(), count)
	} else {
		require.False(t, notifications.Exists())
	}
}

// will be used once the tests that depend on the items structure are adapted to the new test format
func assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t *testing.T, notification gjson.Result, op, formationID, expectedAppID, expectedLocalTenantID, expectedAppNamespace, expectedAppRegion, expectedTenant, expectedCustomerID string, expectedConfig *string) {
	require.Equal(t, op, notification.Get("Operation").String())
	if op == unassignOperation {
		require.Equal(t, expectedAppID, notification.Get("ApplicationID").String())
	}
	require.Equal(t, formationID, notification.Get("RequestBody.ucl-formation-id").String())
	require.Equal(t, expectedTenant, notification.Get("RequestBody.globalAccountId").String())
	require.Equal(t, expectedCustomerID, notification.Get("RequestBody.crmId").String())

	notificationItems := notification.Get("RequestBody.items")
	require.True(t, notificationItems.Exists())
	require.Len(t, notificationItems.Array(), 1)

	app1FromNotification := notificationItems.Array()[0]
	require.Equal(t, expectedAppID, app1FromNotification.Get("ucl-system-tenant-id").String())
	require.Equal(t, expectedLocalTenantID, app1FromNotification.Get("tenant-id").String())
	require.Equal(t, expectedAppNamespace, app1FromNotification.Get("application-namespace").String())
	require.Equal(t, expectedAppRegion, app1FromNotification.Get("region").String())
	if expectedConfig != nil {
		require.Equal(t, *expectedConfig, notification.Get("RequestBody.config").String())
	}
}

func assertTrustDetailsForTargetAndNoTrustDetailsForSource(t *testing.T, assignNotificationAboutApp2 gjson.Result, expectedSubjects []string) {
	t.Logf("Assert trust details are send to the target")
	notificationItems := assignNotificationAboutApp2.Get("RequestBody.items")
	app1FromNotification := notificationItems.Array()[0]
	targetTrustDetails := app1FromNotification.Get("target-trust-details")
	certificateDetails := targetTrustDetails.Array()[0].String()
	certificateDetailsSecond := targetTrustDetails.Array()[1].String()
	sortedSubjects := make([]string, 0, len(expectedSubjects))
	for _, subject := range expectedSubjects {
		sortedSubjects = append(sortedSubjects, certs.SortSubject(subject))
	}
	require.ElementsMatch(t, sortedSubjects, []string{certificateDetails, certificateDetailsSecond})

	t.Logf("Assert that there are no trust details for the source")
	sourceTrustDetails := app1FromNotification.Get("source-trust-details")
	require.Equal(t, 0, len(sourceTrustDetails.Array()))
}

func verifyFormationAssignmentNotification(t *testing.T, notification gjson.Result, op, formationID, expectedObjectID, expectedAppLocalTenantID, expectedObjectNamespace, expectedObjectRegion string, expectedConfiguration *string, expectedTenant, expectedCustomerID string, shouldRemoveDestinationCertificateData bool, expectedState *string) error {
	actualOp := notification.Get("Operation").String()
	if op != actualOp {
		return errors.Errorf("Operation does not match - expected: %q, but got: %q", op, actualOp)
	}

	if op == unassignOperation {
		if actualObjectIDExists := notification.Get("ApplicationID").Exists(); !actualObjectIDExists {
			return errors.New("ObjectID does not exist")
		}

		actualObjectID := notification.Get("ApplicationID").String()
		if expectedObjectID != actualObjectID {
			return errors.Errorf("ObjectID does not match - expected: %q, but got: %q", expectedObjectID, actualObjectID)
		}
	}

	actualFormationID := notification.Get("RequestBody.context.uclFormationId").String()
	if formationID != actualFormationID {
		return errors.Errorf("RequestBody.context.uclFormationId does not match - expected: %q, but got: %q", formationID, actualFormationID)
	}

	actualTenantID := notification.Get("RequestBody.context.globalAccountId").String()
	if expectedTenant != actualTenantID {
		return errors.Errorf("RequestBody.context.globalAccountId does not match - expected: %q, but got: %q", expectedTenant, actualTenantID)
	}

	actualCustomerID := notification.Get("RequestBody.context.crmId").String()
	if expectedCustomerID != actualCustomerID {
		return errors.Errorf("RequestBody.context.crmId does not match - expected: %q, but got: %q", expectedCustomerID, actualCustomerID)
	}

	actualAppTenantID := notification.Get("RequestBody.receiverTenant.applicationTenantId").String()
	if expectedAppLocalTenantID != actualAppTenantID {
		return errors.Errorf("RequestBody.receiverTenant.applicationTenantId does not match - expected: %q, but got: %q", expectedAppLocalTenantID, actualAppTenantID)
	}

	actualObjectRegion := notification.Get("RequestBody.receiverTenant.deploymentRegion").String()
	if expectedObjectRegion != actualObjectRegion {
		return errors.Errorf("RequestBody.receiverTenant.deploymentRegion does not match - expected: %q, but got: %q", expectedObjectRegion, actualObjectRegion)
	}

	actualObjectNamespace := notification.Get("RequestBody.receiverTenant.applicationNamespace").String()
	if expectedObjectNamespace != actualObjectNamespace {
		return errors.Errorf("RequestBody.receiverTenant.applicationNamespace does not match - expected: %q, but got: %q", expectedObjectNamespace, actualObjectNamespace)
	}

	if expectedState != nil {
		notificationReceiverState := notification.Get("RequestBody.receiverTenant.state").String()
		if notificationReceiverState != *expectedState {
			return errors.Errorf("RequestBody.receiverTenant.state does not match - expeted: %q, but got: %q", *expectedState, notificationReceiverState)
		}
	}

	if shouldRemoveDestinationCertificateData {
		notificationReceiverCfg := notification.Get("RequestBody.receiverTenant.configuration").String()
		notificationReceiverState := notification.Get("RequestBody.receiverTenant.state").String()
		if notificationReceiverCfg == "" && notificationReceiverState == "INITIAL" {
			return nil
		}

		modifiedNotification, err := sjson.Delete(notification.String(), "RequestBody.receiverTenant.configuration.credentials.inboundCommunication.samlAssertion.certificate")
		if err != nil {
			return err
		}

		modifiedNotification, err = sjson.Delete(modifiedNotification, "RequestBody.receiverTenant.configuration.credentials.inboundCommunication.clientCertificateAuthentication.certificate")
		if err != nil {
			return err
		}

		modifiedNotification, err = sjson.Delete(modifiedNotification, "RequestBody.receiverTenant.configuration.credentials.inboundCommunication.samlAssertion.assertionIssuer")
		if err != nil {
			return err
		}

		modifiedConfig := gjson.Get(modifiedNotification, "RequestBody.receiverTenant.configuration").String()
		if expectedConfiguration != nil {
			assert.JSONEq(t, *expectedConfiguration, modifiedConfig, "RequestBody.receiverTenant.configuration does not match")
		}
	} else {
		actualConfiguration := notification.Get("RequestBody.receiverTenant.configuration").String()
		if expectedConfiguration != nil && *expectedConfiguration != "" && *expectedConfiguration != actualConfiguration {
			return errors.Errorf("RequestBody.receiverTenant.configuration does not match - expected: %q, but got: %q", *expectedConfiguration, actualConfiguration)
		}
	}

	return nil
}
