package installation

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

const (
	ClusterServiceBrokerNameLabel        = "servicecatalog.k8s.io/spec.clusterServiceBrokerName"
	ClusterServiceClassExternalNameLabel = "servicecatalog.k8s.io/spec.externalName"
	ClusterServiceClassRefNameLabel      = "servicecatalog.k8s.io/spec.clusterServiceClassRef.name"
)

type ServiceCatalogClient interface {
	PerformCleanup(resourceSelector string) error
	listClusterServiceBroker(metav1.ListOptions) (*v1beta1.ClusterServiceBrokerList, error)
	listClusterServiceClass(metav1.ListOptions) (*v1beta1.ClusterServiceClassList, error)
	listServiceInstance(metav1.ListOptions) (*v1beta1.ServiceInstanceList, error)
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

func (s *serviceCatalogClient) PerformCleanup(resourceSelector string) error {
	clusterServiceBrokers, err := s.listClusterServiceBroker(metav1.ListOptions{})
	if err != nil {
		return errors.Wrapf(err, "while listing ClusterServiceBrokers")
	}

	logrus.Debugf("Filtering ClusterServiceBrokers with url prefix %s", resourceSelector)
	brokersWithUrlPrefix := s.filterCsbWithUrlPrefix(clusterServiceBrokers, resourceSelector)
	logrus.Debugf("Got %d ClusterServiceBrokers to process", len(brokersWithUrlPrefix))

	cscWithMatchingLabel, err := s.getClusterServiceClassesForBrokers(brokersWithUrlPrefix)
	if err != nil {
		return errors.Wrapf(err, "while getting ClusterServiceClasses")
	}

	serviceInstances, err := s.getServiceInstancesForClusterServiceClasses(cscWithMatchingLabel)
	if err != nil {
		return errors.Wrapf(err, "while getting ServiceInstances")
	}

	err = s.deleteServiceInstances(serviceInstances)
	if err != nil {
		return errors.Wrapf(err, "while deleting ServiceInstances")
	}

	return nil
}

func (s *serviceCatalogClient) listClusterServiceBroker(options metav1.ListOptions) (*v1beta1.ClusterServiceBrokerList, error) {
	result := &v1beta1.ClusterServiceBrokerList{}
	logrus.Debugf("trying to list ClusterServiceBrokers...")
	err := wait.Poll(10*time.Second, 2*time.Minute, func() (done bool, err error) {
		csbList, err := s.client.ServicecatalogV1beta1().ClusterServiceBrokers().List(options)
		if err != nil {
			logrus.Errorf("while listing ClusterServiceBrokers: %s", err.Error())
			if apiErrors.IsNotFound(err) {
				return true, nil
			}
			return false, nil
		}
		result = csbList
		return true, nil
	})
	return result, err
}

func (s *serviceCatalogClient) listClusterServiceClass(options metav1.ListOptions) (*v1beta1.ClusterServiceClassList, error) {
	result := &v1beta1.ClusterServiceClassList{}
	logrus.Debugf("trying to list ClusterServiceClasses...")
	err := wait.Poll(10*time.Second, 2*time.Minute, func() (done bool, err error) {
		cscList, err := s.client.ServicecatalogV1beta1().ClusterServiceClasses().List(options)
		if err != nil {
			logrus.Errorf("while listing ClusterServiceClasses: %s", err.Error())
			if apiErrors.IsNotFound(err) {
				result = nil
				return true, nil
			}
			return false, nil
		}
		result = cscList
		return true, nil
	})
	return result, err
}

func (s *serviceCatalogClient) listServiceInstance(options metav1.ListOptions) (*v1beta1.ServiceInstanceList, error) {
	result := &v1beta1.ServiceInstanceList{}
	logrus.Debugf("trying to list ServiceInstances...")
	err := wait.Poll(10*time.Second, 2*time.Minute, func() (done bool, err error) {
		siList, err := s.client.ServicecatalogV1beta1().ServiceInstances(metav1.NamespaceAll).List(options)
		if err != nil {
			logrus.Errorf("while listing ServiceInstances: %s", err.Error())
			if apiErrors.IsNotFound(err) {
				result = nil
				return true, nil
			}
			return false, nil
		}
		result = siList
		return true, nil
	})
	return result, err
}

func (s *serviceCatalogClient) filterCsbWithUrlPrefix(csbList *v1beta1.ClusterServiceBrokerList, urlPrefix string) []v1beta1.ClusterServiceBroker {
	var csbWithBrokerUrlPrefix []v1beta1.ClusterServiceBroker
	for _, clusterServiceBroker := range csbList.Items {
		if strings.HasPrefix(clusterServiceBroker.Spec.URL, urlPrefix) {
			csbWithBrokerUrlPrefix = append(csbWithBrokerUrlPrefix, clusterServiceBroker)
		}
	}

	return csbWithBrokerUrlPrefix
}

func (s *serviceCatalogClient) getClusterServiceClassesForBrokers(brokers []v1beta1.ClusterServiceBroker) ([]v1beta1.ClusterServiceClass, error) {
	var cscWithMatchingLabel []v1beta1.ClusterServiceClass

	for _, csb := range brokers {
		labelValue := GenerateSHA(csb.Name)
		csbListOptions := fixListOptionsWithLabelSelector(ClusterServiceBrokerNameLabel, labelValue)

		clusterServiceClasses, err := s.listClusterServiceClass(csbListOptions)
		if err != nil {
			return []v1beta1.ClusterServiceClass{}, errors.Wrapf(err, "while listing ClusterServiceClasses for ClusterServiceBroker %q", csb.Name)
		}

		for _, serviceClass := range clusterServiceClasses.Items {
			logrus.Debugf("found ClusterServiceClass with label %q: %s", labelValue, serviceClass.Name)
			cscWithMatchingLabel = append(cscWithMatchingLabel, serviceClass)
		}
	}

	return cscWithMatchingLabel, nil
}

func (s *serviceCatalogClient) getServiceInstancesForClusterServiceClasses(serviceClasses []v1beta1.ClusterServiceClass) ([]v1beta1.ServiceInstance, error) {
	var serviceInstances []v1beta1.ServiceInstance

	for _, clusterServiceClass := range serviceClasses {
		labelValue := GenerateSHA(clusterServiceClass.Name)

		options := fixListOptionsWithLabelSelector(ClusterServiceClassRefNameLabel, labelValue)

		serviceInstancesList, err := s.listServiceInstance(options)
		if err != nil {
			return []v1beta1.ServiceInstance{}, errors.Wrapf(err, "while listing ServiceInstances")
		}

		for _, serviceInstance := range serviceInstancesList.Items {
			logrus.Debugf("found ServiceInstance with label %q: %s", labelValue, serviceInstance.Name)
			serviceInstances = append(serviceInstances, serviceInstance)
		}
	}
	return serviceInstances, nil
}

func (s *serviceCatalogClient) deleteServiceInstances(serviceInstances []v1beta1.ServiceInstance) error {
	for _, serviceInstance := range serviceInstances {
		logrus.Debugf("trying to delete ServiceInstance %q", serviceInstance.Name)

		_ = wait.Poll(10*time.Second, 2*time.Minute, func() (done bool, err error) {
			if err := s.client.ServicecatalogV1beta1().ServiceInstances(serviceInstance.Namespace).Delete(serviceInstance.Name, &metav1.DeleteOptions{}); err != nil {
				if apiErrors.IsNotFound(err) {
					return true, nil
				}
				logrus.Errorf("while removing ServiceInstance %s: %s", serviceInstance.Name, err.Error())
				return false, nil
			}
			return true, nil
		})
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
func GenerateSHA(input string) string {
	// TODO: remove this and import from "github.com/kubernetes-sigs/service-catalog/pkg/util" when service-catalog v0.3 is released
	h := sha256.New224()
	_, err := h.Write([]byte(input))
	if err != nil {
		logrus.Errorf("cannot generate SHA224 from string %q: %s", input, err)
		return ""
	}

	return hex.EncodeToString(h.Sum(nil))
}
