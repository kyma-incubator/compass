package gardener

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const (
	ClusterServiceBrokerNameLabel        = "servicecatalog.k8s.io/spec.clusterServiceBrokerName"
	ClusterServiceClassExternalNameLabel = "servicecatalog.k8s.io/spec.externalName"
	ClusterServiceClassRefNameLabel      = "servicecatalog.k8s.io/spec.clusterServiceClassRef.name"
)

type ResourceCleanupTimeouts struct {
	ServiceInstanceCleanupTimeout time.Duration `envconfig:"default=20m"`
}

type ResourceSelectors struct {
	BrokerUrlPrefix string `envconfig:"default=https://service-manager."`
}

//go:generate mockery -name=ResourcesJanitor
type ResourcesJanitor interface {
	CleanUpShootResources(shootName string) error
}

type resourcesJanitor struct {
	enabled           bool
	cleanupTimeouts   ResourceCleanupTimeouts
	resourceSelectors ResourceSelectors
	secretsClient     v1core.SecretInterface
}

func NewResourcesJanitor(
	isEnabled bool,
	cleanupTimeouts ResourceCleanupTimeouts,
	selectors ResourceSelectors,
	secretsClient v1core.SecretInterface,
) (ResourcesJanitor, error) {
	return &resourcesJanitor{
		enabled:           isEnabled,
		cleanupTimeouts:   cleanupTimeouts,
		resourceSelectors: selectors,
		secretsClient:     secretsClient,
	}, nil
}

func (rj *resourcesJanitor) CleanUpShootResources(shootName string) error {
	if !rj.enabled {
		return errors.New("shoot resource janitor disabled")
	}
	kubeconfig, err := KubeconfigForShoot(rj.secretsClient, shootName)
	if err != nil {
		return errors.Wrap(err, "error fetching kubeconfig")
	}
	svcCli, err := NewServiceCatalogClient(kubeconfig)
	if err != nil {
		return errors.Wrapf(err, "while creating k8s client for shoot %q", shootName)
	}

	err = cleanUpShootResourcesWithBrokerUrlPrefix(svcCli, rj.resourceSelectors.BrokerUrlPrefix)
	if err != nil {
		return err
	}

	return nil
}

func cleanUpShootResourcesWithBrokerUrlPrefix(cli ResourceClient, brokerUrlPrefix string) error {
	//1. Fetch all service cluster service brokers
	list, err := cli.ListResource("ClusterServiceBroker", metav1.ListOptions{})
	if err != nil {
		return err
	}
	clusterServiceBrokers := list.(*v1beta1.ClusterServiceBrokerList)

	//2. Filter based on url prefix
	var csbWIthBrokerUrlPrefix []v1beta1.ClusterServiceBroker
	for _, clusterServiceBroker := range clusterServiceBrokers.Items {
		if strings.HasPrefix(clusterServiceBroker.Spec.URL, brokerUrlPrefix) {
			csbWIthBrokerUrlPrefix = append(csbWIthBrokerUrlPrefix, clusterServiceBroker)
		}
	}

	//3. Cluster service broker -> util.GenerateSHA
	var cscWithMatchingLabel []v1beta1.ClusterServiceClass

	for _, csb := range csbWIthBrokerUrlPrefix {
		labelValue := GenerateSHA(csb.Name)
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{ClusterServiceBrokerNameLabel: labelValue},
		}
		listOptions := metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
		//4. Fetch all cluster class based on ^ label
		list, err := cli.ListResource("ClusterServiceClass", listOptions)
		if err != nil {
			fmt.Errorf("while listing resources: %s", err)
		}
		clusterServiceClasses := list.(*v1beta1.ClusterServiceClassList)

		for _, serviceClass := range clusterServiceClasses.Items {
			fmt.Println(fmt.Sprintf("ClusterServiceClass with label %q: %s", labelValue, serviceClass.Name))
			cscWithMatchingLabel = append(cscWithMatchingLabel, serviceClass)
		}
	}
	//6. Get servicecatalog.k8s.io/spec.externalName for each ClusterClass and find corresponding Service Instances

	for _, clusterServiceClass := range cscWithMatchingLabel {
		labelValue := clusterServiceClass.Labels[ClusterServiceClassExternalNameLabel]
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{ClusterServiceClassRefNameLabel: labelValue},
		}
		listOptions := metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}

		list, err := cli.ListResource("ServiceInstance", listOptions)
		if err != nil {
			fmt.Errorf("while listing resources: %s", err)
		}

		serviceInstanceList := list.(*v1beta1.ServiceInstanceList)

		for _, svcInstance := range serviceInstanceList.Items {
			fmt.Println(fmt.Sprintf("trying to delete service instance %q", svcInstance.Name))

			err = wait.Poll(time.Second, 3*time.Minute, func() (bool, error) {
				if err := cli.DeleteResource("ServiceInstance", svcInstance.Name, svcInstance.Namespace, &metav1.DeleteOptions{}); err != nil {
					if apiErrors.IsNotFound(err) {
						return true, nil
					}
					fmt.Println(errors.Wrap(err, "while removing instance").Error())
				}
				return false, nil
			})
		}
	}

	return nil
}

type ResourceClient interface {
	ListResource(resourceKind string, options metav1.ListOptions) (runtime.Object, error)
	DeleteResource(resourceKind, name, namespace string, options *metav1.DeleteOptions) error
}

type serviceCatalogClient struct {
	client *scClient.Clientset
}

func NewServiceCatalogClient(kubeconfig *rest.Config) (*serviceCatalogClient, error) {
	scCli, err := scClient.NewForConfig(kubeconfig)
	if err != nil {
		return &serviceCatalogClient{}, err
	}
	return &serviceCatalogClient{client: scCli}, nil
}

func (svc *serviceCatalogClient) ListResource(resourceKind string, options metav1.ListOptions) (runtime.Object, error) {
	switch resourceKind {
	case "ClusterServiceBroker":
		return svc.client.ServicecatalogV1beta1().ClusterServiceBrokers().List(options)
	case "ClusterServiceClass":
		return svc.client.ServicecatalogV1beta1().ClusterServiceClasses().List(options)
	case "ServiceInstance":
		return svc.client.ServicecatalogV1beta1().ServiceInstances(metav1.NamespaceAll).List(options)
	default:
		return nil, errors.New(fmt.Sprintf("requested resource kind %q is not implemented in resource client", resourceKind))
	}
}

func (svc *serviceCatalogClient) DeleteResource(resourceKind, name, namespace string, options *metav1.DeleteOptions) error {
	switch resourceKind {
	case "ServiceInstance":
		return svc.client.ServicecatalogV1beta1().ServiceInstances(namespace).Delete(name, options)
	default:
		return errors.New(fmt.Sprintf("requested resource kind %q is not implemented in resource client", resourceKind))
	}
}

// GenerateSHA generates the sha224 value from the given string
// the function is used to provide a string length less than 63 characters, this string is used in label of resource
// sha algorithm cannot be changed in the future because of backward compatibles
// TODO: remove this and import from "github.com/kubernetes-sigs/service-catalog/pkg/util" when service-catalog v0.3 is released
func GenerateSHA(input string) string {
	h := sha256.New224()
	_, err := h.Write([]byte(input))
	if err != nil {
		klog.Errorf("cannot generate SHA224 from string %q: %s", input, err)
		return ""
	}

	return hex.EncodeToString(h.Sum(nil))
}
