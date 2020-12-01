package kubernetes

import (
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	PollInterval time.Duration `envconfig:"default=2s"`
	PollTimeout  time.Duration `envconfig:"default=1m"`
	Timeout      time.Duration `envconfig:"default=2m"`
}

func NewKubernetesClientSet(interval, pollingTimeout, timeout time.Duration) (*kubernetes.Clientset, error) {
	kubeConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Println("Failed to read in cluster Config", err)
		log.Println("Trying to initialize with local Config")
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
			log.Printf("Failed to access API Server: %s", err.Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	log.Println("Successfully initialized kubernetes client")
	return kubeClientSet, nil
}
