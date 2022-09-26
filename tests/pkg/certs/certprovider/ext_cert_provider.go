package certprovider

import (
	"context"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ExternalCertProviderConfig struct {
	ExternalClientCertTestSecretName         string  `envconfig:"EXTERNAL_CLIENT_CERT_TEST_SECRET_NAME"`
	ExternalClientCertTestSecretNamespace    string  `envconfig:"EXTERNAL_CLIENT_CERT_TEST_SECRET_NAMESPACE"`
	CertSvcInstanceTestSecretName            string  `envconfig:"CERT_SVC_INSTANCE_TEST_SECRET_NAME"`
	CertSvcInstanceTestRegion2SecretName     string  `envconfig:"CERT_SVC_INSTANCE_TEST_REGION2_SECRET_NAME"`
	ExternalCertCronjobContainerName         string  `envconfig:"EXTERNAL_CERT_CRONJOB_CONTAINER_NAME"`
	ExternalCertTestJobName                  string  `envconfig:"EXTERNAL_CERT_TEST_JOB_NAME"`
	TestExternalCertSubject                  string  `envconfig:"TEST_EXTERNAL_CERT_SUBJECT"`
	TestExternalCertSubjectRegion2           string  `envconfig:"TEST_EXTERNAL_CERT_SUBJECT_REGION2"`
	TestExternalCertCN                       string  `envconfig:"TEST_EXTERNAL_CERT_CN"`
	ExternalClientCertCertKey                string  `envconfig:"APP_EXTERNAL_CLIENT_CERT_KEY"`
	ExternalClientCertKeyKey                 string  `envconfig:"APP_EXTERNAL_CLIENT_KEY_KEY"`
	ExternalClientCertExpectedIssuerLocality *string `envconfig:"-"`
	ExternalClientCertSecretName             string  `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}

func NewExternalCertFromConfig(t *testing.T, ctx context.Context, testConfig ExternalCertProviderConfig) (*rsa.PrivateKey, [][]byte) {
	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	createExtCertJob(t, ctx, k8sClient, testConfig, testConfig.ExternalCertTestJobName) // Create temporary external certificate job which will save the modified client certificate in temporary secret
	defer func() {
		k8s.PrintJobLogs(t, ctx, k8sClient, testConfig.ExternalCertTestJobName, testConfig.ExternalClientCertTestSecretNamespace, testConfig.ExternalCertCronjobContainerName, false)

		k8s.DeleteJob(t, ctx, k8sClient, testConfig.ExternalCertTestJobName, testConfig.ExternalClientCertTestSecretNamespace)
		k8s.DeleteSecret(t, ctx, k8sClient, testConfig.ExternalClientCertTestSecretName, testConfig.ExternalClientCertTestSecretNamespace)
	}()
	k8s.WaitForJobToSucceed(t, ctx, k8sClient, testConfig.ExternalCertTestJobName, testConfig.ExternalClientCertTestSecretNamespace)

	providerExtCrtTestSecret, err := k8sClient.CoreV1().Secrets(testConfig.ExternalClientCertTestSecretNamespace).Get(ctx, testConfig.ExternalClientCertTestSecretName, metav1.GetOptions{})
	require.NoError(t, err)
	providerKeyBytes := providerExtCrtTestSecret.Data[testConfig.ExternalClientCertKeyKey]
	require.NotEmpty(t, providerKeyBytes)
	providerCertChainBytes := providerExtCrtTestSecret.Data[testConfig.ExternalClientCertCertKey]
	require.NotEmpty(t, providerCertChainBytes)

	return certs.ClientCertPair(t, providerCertChainBytes, providerKeyBytes)
}

// createExtCertJob will schedule a temporary kubernetes job from director-external-certificate-rotation-job cronjob
// with replaced certificate subject and secret name so the tests can be executed on real environment with the correct values.
func createExtCertJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, testConfig ExternalCertProviderConfig, jobName string) {
	cronjobName := "director-external-certificate-rotation-job"

	cronjob := k8s.GetCronJob(t, ctx, k8sClient, cronjobName, testConfig.ExternalClientCertTestSecretNamespace)

	// change the secret name and certificate subject
	podContainers := &cronjob.Spec.JobTemplate.Spec.Template.Spec.Containers
	for cIndex := range *podContainers {
		container := &(*podContainers)[cIndex]
		if container.Name == testConfig.ExternalCertCronjobContainerName {
			for eIndex := range container.Env {
				env := &container.Env[eIndex]
				if env.Name == "EXPECTED_ISSUER_LOCALITY" && testConfig.ExternalClientCertExpectedIssuerLocality != nil {
					env.Value = *testConfig.ExternalClientCertExpectedIssuerLocality
				}
				if env.Name == "CLIENT_CERT_SECRET_NAME" {
					env.Value = testConfig.ExternalClientCertTestSecretName
				}
				if env.Name == "CERT_SUBJECT_PATTERN" {
					env.Value = testConfig.TestExternalCertSubject
				}
				if env.Name == "CERT_SVC_CSR_ENDPOINT" || env.Name == "CERT_SVC_CLIENT_ID" || env.Name == "CERT_SVC_OAUTH_URL" || env.Name == "CERT_SVC_OAUTH_CLIENT_CERT" || env.Name == "CERT_SVC_OAUTH_CLIENT_KEY" {
					if testConfig.CertSvcInstanceTestSecretName != "" {
						env.ValueFrom.SecretKeyRef.Name = testConfig.CertSvcInstanceTestSecretName // external certificate credentials used to execute consumer-provider test
					} else if testConfig.CertSvcInstanceTestRegion2SecretName != "" {
						env.ValueFrom.SecretKeyRef.Name = testConfig.CertSvcInstanceTestRegion2SecretName // external certificate credentials used to execute consumer-provider test
					}
				}
			}
			break
		}
	}

	jobDef := &v1.Job{
		Spec: v1.JobSpec{
			Parallelism:             cronjob.Spec.JobTemplate.Spec.Parallelism,
			Completions:             cronjob.Spec.JobTemplate.Spec.Completions,
			ActiveDeadlineSeconds:   cronjob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds,
			BackoffLimit:            cronjob.Spec.JobTemplate.Spec.BackoffLimit,
			Selector:                cronjob.Spec.JobTemplate.Spec.Selector,
			ManualSelector:          cronjob.Spec.JobTemplate.Spec.ManualSelector,
			Template:                cronjob.Spec.JobTemplate.Spec.Template,
			TTLSecondsAfterFinished: cronjob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
	}

	k8s.CreateJobByGivenJobDefinition(t, ctx, k8sClient, jobName, testConfig.ExternalClientCertTestSecretNamespace, jobDef)
}
