package formationnotification_test

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/formationnotification"
)

var (
	testTenantID = "testTenantID"
	url          = "https://target-url.com"

	// Formation Assignments variables
	formationAssignmentReqBody       = `{"ucl-formation-id":"96bbd806-0d56-4f39-bcf1-e15aee9e50cc","items":[{"region":"testRegion","application-namespace":"appNamespce","tenant-id":"localTenantID","ucl-system-tenant-id":"9d8bb1a5-4799-453c-a406-84439f151d45"}]}`
	formationAssignmentReqConfigBody = `{"ucl-formation-id":"96bbd806-0d56-4f39-bcf1-e15aee9e50cc","config": "{\"key\":\"value\"}","items":[{"region":"testRegion","application-namespace":"appNamespce","tenant-id":"localTenantID","ucl-system-tenant-id":"9d8bb1a5-4799-453c-a406-84439f151d45"}]}`
	tenantIDParam                    = "tenantId"
	appID                            = "testAppID"

	// Formation variables
	formationIDParam = "uclFormationId"
	formationReqBody = `{"details":{"id":"cbdd8a33-125f-48a8-b8d9-deebd0b3c168","name":"my-formation"}}`
)

func fixFormationMappings(formationOperation formationnotification.Operation, formationID, formationReqBody string) map[string][]formationnotification.Response {
	return fixMappings(formationOperation, formationID, formationReqBody, nil)
}

func fixFormationAssignmentMappings(formationAssignmentOperation formationnotification.Operation, tenantID, formationAssignmentReqBody string, applicationID *string) map[string][]formationnotification.Response {
	return fixMappings(formationAssignmentOperation, tenantID, formationAssignmentReqBody, applicationID)
}

func fixMappings(operation formationnotification.Operation, objectID, reqBody string, applicationID *string) map[string][]formationnotification.Response {
	return map[string][]formationnotification.Response{
		objectID: {
			{
				Operation:     operation,
				ApplicationID: applicationID,
				RequestBody:   json.RawMessage(reqBody),
			},
		},
	}
}
