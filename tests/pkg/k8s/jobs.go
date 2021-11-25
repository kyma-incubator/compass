package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateJobByCronJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, cronJobName, jobName, namespace string) {
	cronjob, err := k8sClient.BatchV1beta1().CronJobs(namespace).Get(ctx, cronJobName, metav1.GetOptions{})
	require.NoError(t, err)
	t.Logf("Got the cronjob %s", cronJobName)

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
	t.Logf("Creating test job %s", jobName)
	_, err = k8sClient.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Logf("Test job created %s", jobName)
}

func DeleteJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	t.Logf("Deleting test job %s", jobName)

	var gracePeriod int64 = 0
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
	elapsed := time.After(time.Minute * 15)
	for {
		select {
		case <-elapsed:
			t.Fatalf("Timeout reached waiting for job %s to complete. Exiting...", jobName)
		default:
		}
		t.Logf("Waiting for job %s to finish", jobName)
		job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		require.NoError(t, err)
		if job.Status.Failed > 0 {
			t.Fatalf("Job %s has failed. Exiting...", jobName)
		}
		if job.Status.Succeeded > 0 {
			break
		}
		time.Sleep(time.Second * 2)
	}
}
