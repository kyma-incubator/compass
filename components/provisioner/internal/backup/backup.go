package backup

import (
	"fmt"
	"github.com/vmware-tanzu/velero/pkg/apis/velero/v1"

	clientset "github.com/vmware-tanzu/velero/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

//go:generate mockery -name=Service
type Service interface {
	ScheduleBackup(kubeconfig string) error
}

type service struct {
	namespace string
	hour      int
	minutes   int
}

func NewService(namespace string, hour int, minutes int) Service {
	return service{
		namespace: namespace,
		hour:      hour,
		minutes:   minutes,
	}
}

func (cp service) ScheduleBackup(kubeconfig string) error {

	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kubeconfig))
	if err != nil {
		return err
	}

	restConfig, err := clientConfig.ClientConfig()

	clientset, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	includeClusterResources := true
	schedule := v1.Schedule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Schedule",
			APIVersion: "velero.io/v1",
		},
		Spec: v1.ScheduleSpec{
			Template: v1.BackupSpec{
				IncludedNamespaces:      []string{"*"},
				IncludedResources:       []string{"*"},
				IncludeClusterResources: &includeClusterResources,
				StorageLocation:         "default",
				VolumeSnapshotLocations: []string{"default"},
			},
			Schedule: fmt.Sprintf("%d %d * * *", cp.minutes, cp.hour),
		},
	}

	_, err = clientset.VeleroV1().Schedules(cp.namespace).Create(&schedule)

	return err
}
