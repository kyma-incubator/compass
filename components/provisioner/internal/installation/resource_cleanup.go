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
	// Fetch all ClusterServiceBrokers
	//TODO: Add retries for listing resources in case api-server fails to respond
	clusterServiceBrokers, err := s.ListClusterServiceBroker(metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "while listing ClusterServiceBrokers")
	}

	// Filter the list based on ClusterServiceBroker's URL prefix
	brokersWithUrlPrefix := s.FilterCsbWithUrlPrefix(clusterServiceBrokers, BrokerUrlPrefix)

	// Get all ClusterServiceClasses from filtered ClusterServiceBrokers
	cscWithMatchingLabel, err := s.GetClusterServiceClassesForBrokers(brokersWithUrlPrefix)
	if err != nil {
		return errors.Wrapf(err, "while getting ClusterServiceClasses")
	}

	// Get ServiceInstances for each ClusterServiceClass
	serviceInstances, err := s.GetServiceInstancesForClusterServiceClasses(cscWithMatchingLabel)
	if err != nil {
		return errors.Wrapf(err, "while getting ServiceInstances")
	}

	err = s.DeleteServiceInstances(serviceInstances)
	if err != nil {
		return errors.Wrapf(err, "while deleting ServiceInstances")
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

func (s *serviceCatalogClient) FilterCsbWithUrlPrefix(csbList *v1beta1.ClusterServiceBrokerList, urlPrefix string) []v1beta1.ClusterServiceBroker {
	var csbWithBrokerUrlPrefix []v1beta1.ClusterServiceBroker
	for _, clusterServiceBroker := range csbList.Items {
		if strings.HasPrefix(clusterServiceBroker.Spec.URL, urlPrefix) {
			csbWithBrokerUrlPrefix = append(csbWithBrokerUrlPrefix, clusterServiceBroker)
		}
	}

	return csbWithBrokerUrlPrefix
}

func (s *serviceCatalogClient) GetClusterServiceClassesForBrokers(brokers []v1beta1.ClusterServiceBroker) ([]v1beta1.ClusterServiceClass, error) {
	var cscWithMatchingLabel []v1beta1.ClusterServiceClass

	for _, csb := range brokers {
		// Generate label value from ClusterServiceBroker's name
		labelValue := GenerateSHA(csb.Name)
		csbListOptions := fixListOptionsWithLabelSelector(ClusterServiceBrokerNameLabel, labelValue)

		// Fetch all ClusterServiceClasses for single ClusterServiceBroker
		//TODO: Add retries for listing resources in case api-server fails to respond
		clusterServiceClasses, err := s.ListClusterServiceClass(csbListOptions)
		if err != nil {
			return []v1beta1.ClusterServiceClass{}, errors.Wrapf(err, "while listing ClusterServiceClasses for ClusterServiceBroker %q", csb.Name)
		}

		for _, serviceClass := range clusterServiceClasses.Items {
			//TODO: remove this or switch to logger.Debug
			fmt.Println(fmt.Sprintf("found ClusterServiceClass with label %q: %s", labelValue, serviceClass.Name))
			cscWithMatchingLabel = append(cscWithMatchingLabel, serviceClass)
		}
	}

	return cscWithMatchingLabel, nil
}

func (s *serviceCatalogClient) GetServiceInstancesForClusterServiceClasses(serviceClasses []v1beta1.ClusterServiceClass) ([]v1beta1.ServiceInstance, error) {
	var serviceInstances []v1beta1.ServiceInstance

	for _, clusterServiceClass := range serviceClasses {
		// Get ClusterServiceClassExternalName label for single ClusterServiceClass
		labelValue := clusterServiceClass.Labels[ClusterServiceClassExternalNameLabel]

		options := fixListOptionsWithLabelSelector(ClusterServiceClassRefNameLabel, labelValue)

		// Get all ServiceInstances with label referencing to single ClusterServiceClass
		//TODO: Add retries for listing resources in case api-server fails to respond
		serviceInstancesList, err := s.ListServiceInstance(options)
		if err != nil {
			return []v1beta1.ServiceInstance{}, errors.Wrapf(err, "while listing ServiceInstances")
		}

		for _, serviceInstance := range serviceInstancesList.Items {
			//TODO: remove this or switch to logger.Debug
			fmt.Println(fmt.Sprintf("found ServiceInstance with label %q: %s", labelValue, serviceInstance.Name))
			serviceInstances = append(serviceInstances, serviceInstance)
		}
	}
	return serviceInstances, nil
}

func (s *serviceCatalogClient) DeleteServiceInstances(serviceInstances []v1beta1.ServiceInstance) error {
	for _, instance := range serviceInstances {
		//TODO: remove this or switch to logger.Debug
		fmt.Println(fmt.Sprintf("trying to delete ServiceInstance %q", instance.Name))

		err := wait.Poll(10*time.Second, 2*time.Minute, func() (bool, error) {
			if err := s.client.ServicecatalogV1beta1().ServiceInstances(instance.Namespace).Delete(instance.Name, &metav1.DeleteOptions{}); err != nil {
				if apiErrors.IsNotFound(err) {
					return true, nil
				}
				fmt.Println(errors.Wrap(err, "while removing instance").Error())
				return false, nil
			}
			return true, nil
		})

		return err
	}

	return nil
}

func fixListOptionsWithLabelSelector(labelName, labelValue string) metav1.ListOptions {
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{labelName: labelValue},
	}

	return metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
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
