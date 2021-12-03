package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"
	"github.com/stretchr/testify/require"
)

func BuildMockedWebhook(externalSystemURL string, webhookType graphql.WebhookType) *graphql.WebhookInput {
	operationFullPath := BuildOperationFullPath(externalSystemURL)
	deleteFullPath := BuildDeleteFullPath(externalSystemURL)

	return &graphql.WebhookInput{
		Type:           webhookType,
		Mode:           WebhookModePtr(graphql.WebhookModeAsync),
		URLTemplate:    str.Ptr(fmt.Sprintf("{ \\\"method\\\": \\\"DELETE\\\", \\\"path\\\": \\\"%s\\\" }", deleteFullPath)),
		RetryInterval:  IntPtr(5),
		OutputTemplate: str.Ptr(fmt.Sprintf("{ \\\"location\\\": \\\"%s\\\", \\\"success_status_code\\\": 200, \\\"error\\\": \\\"{{.Body.error}}\\\" }", operationFullPath)),
		StatusTemplate: str.Ptr("{ \\\"status\\\": \\\"{{.Body.status}}\\\", \\\"success_status_code\\\": 200, \\\"success_status_identifier\\\": \\\"SUCCEEDED\\\", \\\"in_progress_status_identifier\\\": \\\"IN_PROGRESS\\\", \\\"failed_status_identifier\\\": \\\"FAILED\\\", \\\"error\\\": \\\"{{.Body.error}}\\\" }"),
	}
}

func BuildOperationFullPath(externalSystemURL string) string {
	return fmt.Sprintf("%s%s", externalSystemURL, "webhook/delete/operation")
}

func BuildDeleteFullPath(externalSystemURL string) string {
	return fmt.Sprintf("%s%s", externalSystemURL, "webhook/delete")
}

func UnlockWebhook(t *testing.T, operationFullPath string) {
	httpClient := http.Client{}
	requestData := webhook.OperationStatusRequestData{
		InProgress: false,
	}
	jsonRequestData, err := json.Marshal(requestData)
	require.NoError(t, err)
	reqPost, err := http.NewRequest(http.MethodPost, operationFullPath, bytes.NewBuffer(jsonRequestData))
	require.NoError(t, err)
	respPost, err := httpClient.Do(reqPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respPost.StatusCode)
}

func WebhookModePtr(mode graphql.WebhookMode) *graphql.WebhookMode {
	return &mode
}

func IntPtr(n int) *int {
	return &n
}
