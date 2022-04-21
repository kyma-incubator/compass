package kubernetes

import (
	"context"
	"path/filepath"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Config missing godoc
type Config struct {
	PollInterval time.Duration `envconfig:"optional,default=2s,APP_KUBERNETES_POLL_INTERVAL"`
	PollTimeout  time.Duration `envconfig:"optional,default=1m,APP_KUBERNETES_POLL_TIMEOUT"`
	Timeout      time.Duration `envconfig:"optional,default=2m,APP_KUBERNETES_TIMEOUT"`
}

// NewKubernetesClientSet missing godoc
func NewKubernetesClientSet(ctx context.Context, interval, pollingTimeout, timeout time.Duration) (*kubernetes.Clientset, error) {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while trying to read in cluster Config: %v", err)
		log.C(ctx).Debug("Trying to initialize Kubernetes Client with local Config")
		home := homedir.HomeDir()
		kubeConfPath := filepath.Join(home, ".kube", "Config")
		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfPath)
		if err != nil {
			return nil, errors.Errorf("failed to read k8s in-cluster configuration, %s", err.Error())
		}
	}

	kubeConfig.Timeout = timeout

	kubeClientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	err = wait.PollImmediate(interval, pollingTimeout, func() (bool, error) {
		_, err := kubeClientSet.ServerVersion()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while trying to access API Server: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	log.C(ctx).Info("Successfully initialized kubernetes client")
	return kubeClientSet, nil
}
