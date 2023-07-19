package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
	t.Logf("Got the cronjob %q from %q namespace", cronJobName, namespace)
	return cronjob
}

func CreateJobByGivenJobDefinition(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string, job *v1.Job) {
	t.Logf("Creating test job with name: %q...", jobName)
	_, err := k8sClient.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Logf("Test job with name %q was successfully created", jobName)
}

func DeleteSecret(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, secretName, namespace string) {
	t.Logf("Deleting test secret %q in %q namespace...", secretName, namespace)
	err := k8sClient.CoreV1().Secrets(namespace).Delete(ctx, secretName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod, PropagationPolicy: nil})
	if err != nil && strings.Contains(err.Error(), "not found") {
		require.Error(t, err)
		t.Logf("Test secret %q in %q namespace does not exists", secretName, namespace)
		return
	}

	require.NoError(t, err)
	t.Logf("Test secret %q in %q namespace was successfully deleted", secretName, namespace)
}

func DeleteJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	t.Logf("Deleting test job with name: %q...", jobName)

	propagationPolicy := metav1.DeletePropagationForeground
	err := k8sClient.BatchV1().Jobs(namespace).Delete(ctx, jobName, metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod, PropagationPolicy: &propagationPolicy})

	require.NoError(t, err)

	elapsed := time.After(time.Minute * 3)
	pendingDelete := time.After(time.Minute * 2)
	for {
		select {
		case <-pendingDelete:
			t.Logf("Removing finalizers from the job %q...", jobName)
			finalizers := struct {
				Metadata struct {
					Finalizers interface{} `json:"finalizers"`
				} `json:"metadata"`
			}{
				Metadata: struct {
					Finalizers interface{} `json:"finalizers"`
				}{
					Finalizers: nil,
				},
			}

			patchBytes, err := json.Marshal(finalizers)
			if err != nil {
				t.Fatalf("Can't marshal patch bytes for job %q: %v. Exiting...", jobName, err)
			}

			if _, err = k8sClient.BatchV1().Jobs(namespace).Patch(ctx, jobName, types.MergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
				spew.Dump(err)
				t.Fatalf("Can't patch job %q finalizers: %v. Exiting...", jobName, err)
			}
		case <-elapsed:
			t.Fatalf("Timeout reached waiting for job %q to be deleted. Exiting...", jobName)
		default:
		}
		t.Logf("Waiting for job %q to be deleted...", jobName)
		if _, err = k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{}); errors.IsNotFound(err) {
			break
		}
		time.Sleep(time.Second * 5)
	}
	t.Logf("Test job with name %q was successfully deleted", jobName)
}

func WaitForJobToSucceed(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	WaitForJob(t, ctx, k8sClient, jobName, namespace, false)
}

func WaitForJobToFail(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	WaitForJob(t, ctx, k8sClient, jobName, namespace, true)
}

func WaitForJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string, shouldFail bool) {
	status := getJobStatus(t, ctx, k8sClient, jobName, namespace)
	if status.Failed > 0 {
		if !shouldFail {
			t.Fatalf("Job %q has failed while expecting to succeed. Exiting...", jobName)
		} else {
			return
		}
	}
	if status.Succeeded > 0 {
		if shouldFail {
			t.Fatalf("Job %q has succeeded while expecting to fail. Exiting...", jobName)
		} else {
			return
		}
	}
}

func PrintJobLogs(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace, containerName string, shouldJobFail bool) {
	if shouldJobFail {
		return
	}

	labelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"job-name": jobName}}
	listOptions := metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}
	podsList, err := k8sClient.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		t.Errorf("Failed to list pods for job %q in namespace %q: %s", jobName, namespace, err)
		return
	}

	for _, pod := range podsList.Items {
		if pod.Status.Phase != corev1.PodFailed {
			continue
		}

		t.Logf("Job %q pod %q logs...\n", jobName, pod.Name)

		req := k8sClient.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: containerName})

		podLogs, err := req.Stream(ctx)
		if err != nil {
			t.Errorf("Failed to get logs from pod %q: %s", pod.Name, err)
			return
		}

		buf := new(bytes.Buffer)
		if _, err = io.Copy(buf, podLogs); err != nil {
			t.Errorf("Failed to copy logs from pod %q into buffer: %s", pod.Name, err)
			return
		}

		t.Log(buf.String())

		if err := podLogs.Close(); err != nil {
			t.Errorf("Failed to close %q logs body: %s", pod.Name, err)
			return
		}
	}
}

func getJobStatus(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) v1.JobStatus {
	elapsed := time.After(time.Minute * 15)
	for {
		select {
		case <-elapsed:
			t.Fatalf("Timeout reached waiting for job %q to complete. Exiting...", jobName)
		default:
		}
		t.Logf("Waiting for job %q to finish...", jobName)
		job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		require.NoError(t, err)
		if job.Status.Failed > 0 || job.Status.Succeeded > 0 {
			return job.Status
		}
		time.Sleep(time.Second * 5)
	}
	return v1.JobStatus{}
}
