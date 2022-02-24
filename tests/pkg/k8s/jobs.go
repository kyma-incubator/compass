package k8s

import (
	"context"
	"testing"
	"time"

	"k8s.io/api/batch/v1beta1"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var gracePeriod int64 = 0

func CreateJobByCronJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, cronJobName, jobName, namespace string) {
	cronjob := GetCronJob(t, ctx, k8sClient, cronJobName, namespace)

	job := &v1.Job{
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

	CreateJobByGivenJobDefinition(t, ctx, k8sClient, jobName, namespace, job)
}

func GetCronJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, cronJobName, namespace string) *v1beta1.CronJob {
	cronjob, err := k8sClient.BatchV1beta1().CronJobs(namespace).Get(ctx, cronJobName, metav1.GetOptions{})
	require.NoError(t, err)
	t.Logf("Got the cronjob \"%s\" from \"%s\" namespace", cronJobName, namespace)
	return cronjob
}

func CreateJobByGivenJobDefinition(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string, job *v1.Job) {
	t.Logf("Creating test job with name: %s", jobName)
	_, err := k8sClient.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Logf("Test job with name %s was successfully created", jobName)
}

func DeleteSecret(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, secretName, namespace string) {
	t.Logf("Deleting test secret %q in %q namespace", secretName, namespace)
	err := k8sClient.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod, PropagationPolicy: nil})
	require.NoError(t, err)
	t.Logf("Test secret %q in %q namespace was successfully deleted", secretName, namespace)
}

func DeleteJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	t.Logf("Deleting test job %s", jobName)

	propagationPolicy := metav1.DeletePropagationForeground
	err := k8sClient.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod, PropagationPolicy: &propagationPolicy})

	require.NoError(t, err)

	elapsed := time.After(time.Minute * 2)
	for {
		select {
		case <-elapsed:
			t.Fatalf("Timeout reached waiting for job %s to be deleted. Exiting...", jobName)
		default:
		}
		t.Logf("Waiting for job %s to be deleted", jobName)
		_, err = k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			break
		}
		time.Sleep(time.Second * 2)
	}
	t.Logf("Test job %s deleted", jobName)
}

func WaitForJobToSucceed(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	WaitForJob(t, ctx, k8sClient, jobName, namespace, false)
}

func WaitForJobToFail(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	WaitForJob(t, ctx, k8sClient, jobName, namespace, true)
}

func WaitForJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string, shouldFail bool) {
	elapsed := time.After(time.Minute * 15)
	for {
		select {
		case <-elapsed:
			t.Fatalf("Timeout reached waiting for job %s to complete. Exiting...", jobName)
		default:
		}
		t.Logf("Waiting for job %s to finish...", jobName)
		job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		require.NoError(t, err)
		if job.Status.Failed > 0 {
			if !shouldFail {
				t.Fatalf("Job %s has failed while expecting to succeed. Exiting...", jobName)
			} else {
				break
			}
		}
		if job.Status.Succeeded > 0 {
			if shouldFail {
				t.Fatalf("Job %s has succeeded while expecting to fail. Exiting...", jobName)
			} else {
				break
			}
		}
		time.Sleep(time.Second * 2)
	}
}
