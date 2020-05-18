package installation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

const (
	BrokerUrlPrefix                      = "https://service-manager."
	ClusterServiceBrokerNameLabel        = "servicecatalog.k8s.io/spec.clusterServiceBrokerName"
	ClusterServiceClassExternalNameLabel = "servicecatalog.k8s.io/spec.externalName"
	ClusterServiceClassRefNameLabel      = "servicecatalog.k8s.io/spec.clusterServiceClassRef.name"
)

type ResourceClient interface {
	PerformCleanup() error
}

type ServiceCatalogClient interface {
	PerformCleanup() error
	ListClusterServiceBroker(metav1.ListOptions) (*v1beta1.ClusterServiceBrokerList, error)
	ListClusterServiceClass(metav1.ListOptions) (*v1beta1.ClusterServiceClassList, error)
	ListServiceInstance(metav1.ListOptions) (*v1beta1.ServiceInstanceList, error)
}

func NewServiceCatalogClient(kubeconfig *rest.Config) (ServiceCatalogClient, error) {
	scCli, err := sc.NewForConfig(kubeconfig)
	if err != nil {
		return &serviceCatalogClient{}, err
	}

	return &serviceCatalogClient{client: scCli}, nil
}

type serviceCatalogClient struct {
	client sc.Interface
}

func (s *serviceCatalogClient) PerformCleanup() error {
	//1. Fetch all service cluster service brokers
	clusterServiceBrokers, err := s.ListClusterServiceBroker(metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "while listing cluster service brokers")
	}

	//2. Filter based on broker URL prefix
	var csbWithBrokerUrlPrefix []v1beta1.ClusterServiceBroker
	for _, clusterServiceBroker := range clusterServiceBrokers.Items {
		if strings.HasPrefix(clusterServiceBroker.Spec.URL, BrokerUrlPrefix) {
			csbWithBrokerUrlPrefix = append(csbWithBrokerUrlPrefix, clusterServiceBroker)
		}
	}

	//3. Cluster service broker -> util.GenerateSHA
	var cscWithMatchingLabel []v1beta1.ClusterServiceClass

	for _, csb := range csbWithBrokerUrlPrefix {
		labelValue := GenerateSHA(csb.Name)
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{ClusterServiceBrokerNameLabel: labelValue},
		}
		listOptions := metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
		//4. Fetch all cluster class based on ^ label
		clusterServiceClasses, err := s.ListClusterServiceClass(listOptions)
		if err != nil {
			return errors.Wrapf(err, "while listing cluster service classes")
		}

		for _, serviceClass := range clusterServiceClasses.Items {
			fmt.Println(fmt.Sprintf("ClusterServiceClass with label %q: %s", labelValue, serviceClass.Name))
			cscWithMatchingLabel = append(cscWithMatchingLabel, serviceClass)
		}
	}

	//6. Get ClusterServiceClassExternalNameLabel for each ClusterClass and find corresponding Service Instances
	for _, clusterServiceClass := range cscWithMatchingLabel {
		labelValue := clusterServiceClass.Labels[ClusterServiceClassExternalNameLabel]
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{ClusterServiceClassRefNameLabel: labelValue},
		}
		listOptions := metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}

		serviceInstances, err := s.ListServiceInstance(listOptions)
		if err != nil {
			return errors.Wrapf(err, "while listing service instances")
		}

		for _, instance := range serviceInstances.Items {
			fmt.Println(fmt.Sprintf("trying to delete service instance %q", instance.Name))

			err = wait.Poll(10*time.Second, 3*time.Minute, func() (bool, error) {
				if err := s.client.ServicecatalogV1beta1().ServiceInstances(instance.Namespace).Delete(instance.Name, &metav1.DeleteOptions{}); err != nil {
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

func (s *serviceCatalogClient) ListClusterServiceBroker(options metav1.ListOptions) (*v1beta1.ClusterServiceBrokerList, error) {
	return s.client.ServicecatalogV1beta1().ClusterServiceBrokers().List(options)
}

func (s *serviceCatalogClient) ListClusterServiceClass(options metav1.ListOptions) (*v1beta1.ClusterServiceClassList, error) {
	return s.client.ServicecatalogV1beta1().ClusterServiceClasses().List(options)
}

func (s *serviceCatalogClient) ListServiceInstance(options metav1.ListOptions) (*v1beta1.ServiceInstanceList, error) {
	return s.client.ServicecatalogV1beta1().ServiceInstances(metav1.NamespaceAll).List(options)
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
