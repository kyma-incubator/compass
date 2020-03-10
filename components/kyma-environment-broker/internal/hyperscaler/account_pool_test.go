package hyperscaler

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	machineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHyperscalerTypeFromProviderString(t *testing.T) {

	var testcases = []struct {
		providerType string
		recognized   bool
	}{
		{"GCP", true},
		{"Gcp", true},
		{"gcp", true},
		{"gcpx", false},
		{"Azure", true},
		{"AWS", true},
		{"FOO", false},
	}
	for _, testcase := range testcases {
		var testDescription string
		var expectedError string
		if testcase.recognized {
			testDescription = fmt.Sprintf("%s is a recognized HyperscalerType", testcase.providerType)
			expectedError = ""
		} else {
			testDescription = fmt.Sprintf("%s is an unknown HyperscalerType", testcase.providerType)
			expectedError = fmt.Sprintf("Unknown Hyperscaler provider type: %s", testcase.providerType)
		}
		t.Run(testDescription, func(t *testing.T) {

			hyperscalerType, err := HyperscalerTypeFromProviderString(testcase.providerType)

			if err == nil {
				assert.Equal(t, true, strings.EqualFold(testcase.providerType, string(hyperscalerType)), testDescription)
				assert.Equal(t, expectedError, "", testDescription)
			} else {
				assert.Equal(t, expectedError, err.Error(), testDescription)
			}
		})
	}
}

func TestCredentials(t *testing.T) {

	pool := newTestAccountPool()

	var testcases = []struct {
		testDescription        string
		tenantName             string
		hyperscalerType        HyperscalerType
		expectedCredentialName string
		expectedError          string
	}{
		{"In-use credential for tenant1, GCP returns existing secret",
			"tenant1", GCP, "secret1", ""},

		{"In-use credential for tenant1, Azure returns existing secret",
			"tenant1", Azure, "secret2", ""},

		{"In-use credential for tenant2, GCP returns existing secret",
			"tenant2", GCP, "secret3", ""},

		{"Available credential for tenant3, AWS labels and returns existing secret",
			"tenant3", GCP, "secret4", ""},

		{"Available credential for tenant4, GCP labels and returns existing secret",
			"tenant4", AWS, "secret5", ""},

		{"No Available credential for tenant5, Azure returns error",
			"tenant5", Azure, "",
			"AccountPool failed to find unassigned secret for hyperscalerType: azure"},
	}
	for _, testcase := range testcases {

		t.Run(testcase.testDescription, func(t *testing.T) {

			credentials, err := pool.Credentials(testcase.hyperscalerType, testcase.tenantName)
			actualError := ""
			if err != nil {
				actualError = err.Error()
				assert.Equal(t, testcase.expectedError, actualError)
			} else {
				assert.Equal(t, testcase.expectedCredentialName, credentials.CredentialName)
				assert.Equal(t, testcase.hyperscalerType, credentials.HyperscalerType)
				assert.Equal(t, testcase.tenantName, credentials.TenantName)
				assert.Equal(t, testcase.expectedCredentialName, string(credentials.CredentialData["credentials"]))
				assert.Equal(t, testcase.expectedError, actualError)
			}
		})
	}
}

func newTestAccountPool() AccountPool {
	var testNamespace = "test-namespace"

	secret1 := &corev1.Secret{
		ObjectMeta: machineryv1.ObjectMeta{
			Name: "secret1", Namespace: testNamespace,
			Labels: map[string]string{
				"tenantName":      "tenant1",
				"hyperscalerType": "gcp",
			},
		},
		Data: map[string][]byte{
			"credentials": []byte("secret1"),
		},
	}
	secret2 := &corev1.Secret{
		ObjectMeta: machineryv1.ObjectMeta{
			Name: "secret2", Namespace: testNamespace,
			Labels: map[string]string{
				"tenantName":      "tenant1",
				"hyperscalerType": "azure",
			},
		},
		Data: map[string][]byte{
			"credentials": []byte("secret2"),
		},
	}
	secret3 := &corev1.Secret{
		ObjectMeta: machineryv1.ObjectMeta{
			Name: "secret3", Namespace: testNamespace,
			Labels: map[string]string{
				"tenantName":      "tenant2",
				"hyperscalerType": "gcp",
			},
		},
		Data: map[string][]byte{
			"credentials": []byte("secret3"),
		},
	}
	secret4 := &corev1.Secret{
		ObjectMeta: machineryv1.ObjectMeta{
			Name: "secret4", Namespace: testNamespace,
			Labels: map[string]string{
				"hyperscalerType": "gcp",
			},
		},
		Data: map[string][]byte{
			"credentials": []byte("secret4"),
		},
	}
	secret5 := &corev1.Secret{
		ObjectMeta: machineryv1.ObjectMeta{
			Name: "secret5", Namespace: testNamespace,
			Labels: map[string]string{
				"hyperscalerType": "aws",
			},
		},
		Data: map[string][]byte{
			"credentials": []byte("secret5"),
		},
	}

	mockClient := fake.NewSimpleClientset(secret1, secret2, secret3, secret4, secret5)
	mockSecrets := mockClient.CoreV1().Secrets(testNamespace)
	pool := NewAccountPool(mockSecrets)
	return pool
}
