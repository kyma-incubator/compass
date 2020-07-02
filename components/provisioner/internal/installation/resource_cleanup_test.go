package installation

import (
	"testing"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scfake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kubernetes-sigs/service-catalog/pkg/util"
	"github.com/kyma-project/control-plane/components/provisioner/internal/util/k8s"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgotesting "k8s.io/client-go/testing"
)

const (
	TestResourceSelector = "https://service-manager."

	ClusterBrokerNameProvidingClusterServiceClasses    = "sm-working-broker"
	ClusterBrokerNameNotProvidingClusterServiceClasses = "sm-fake-broker"

	ClusterServiceClassFixedName = "some-serviceclass-name"
)

func TestNewServiceCatalogClient(t *testing.T) {
	//given
	k8sConfig, err := k8s.ParseToK8sConfig([]byte(kubeconfig))
	require.NoError(t, err)

	t.Run("should create client with valid kubeconfig", func(t *testing.T) {
		// when
		cli, err := NewServiceCatalogClient(k8sConfig)

		//then
		require.NoError(t, err)
		require.NotNil(t, cli)
	})
}

func TestServiceCatalogClient_PerformCleanup(t *testing.T) {
	// given
	fakeClient := scfake.NewSimpleClientset(newTestCR()...)
	cli := &serviceCatalogClient{client: fakeClient}

	t.Run("should perform resource cleanup succesfully with happy path", func(t *testing.T) {
		// when
		err := cli.PerformCleanup(TestResourceSelector)

		// then
		require.NoError(t, err)

		// when
		res, err := cli.listServiceInstance(metav1.ListOptions{})
		require.NoError(t, err)
		assert.Nil(t, res.Items)
	})

	t.Run("should fail cleanup when unable to list ClusterServiceBrokers", func(t *testing.T) {
		// given
		fakeClient := scfake.NewSimpleClientset(newTestCR()...)
		cli := &serviceCatalogClient{client: fakeClient}
		fakeClient.PrependReactor("list", "clusterservicebrokers", func(action clientgotesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("error listing clusterservicebrokers")
		})

		// when
		err := cli.PerformCleanup(TestResourceSelector)

		// then
		require.Error(t, err)
	})

	t.Run("should fail cleanup when unable to list ClusterServiceClasses", func(t *testing.T) {
		// given
		fakeClient := scfake.NewSimpleClientset(newTestCR()...)
		cli := &serviceCatalogClient{client: fakeClient}
		fakeClient.PrependReactor("list", "clusterserviceclasses", func(action clientgotesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("error listing clusterserviceclasses")
		})

		// when
		err := cli.PerformCleanup(TestResourceSelector)

		// then
		require.Error(t, err)
	})

	t.Run("should fail cleanup when unable to list ServiceInstances", func(t *testing.T) {
		// given
		fakeClient := scfake.NewSimpleClientset(newTestCR()...)
		cli := &serviceCatalogClient{client: fakeClient}
		fakeClient.PrependReactor("list", "serviceinstances", func(action clientgotesting.Action) (bool, runtime.Object, error) {
			return true, nil, errors.New("error listing serviceinstances")
		})

		// when
		err := cli.PerformCleanup(TestResourceSelector)

		// then
		require.Error(t, err)
	})
}

func TestServiceCatalogClient_ListClusterServiceBroker(t *testing.T) {
	// given
	fakeClient := scfake.NewSimpleClientset(newTestCR()...)
	cli := &serviceCatalogClient{client: fakeClient}

	t.Run("should list ClusterServiceBrokers successfully", func(t *testing.T) {
		// when
		list, err := cli.listClusterServiceBroker(metav1.ListOptions{})

		// then
		require.NoError(t, err)
		require.NotNil(t, list.Items)
		assert.Len(t, list.Items, 3)
	})

	//given
	fakeClient = scfake.NewSimpleClientset()
	cli = &serviceCatalogClient{client: fakeClient}

	t.Run("should return nil if no ClusterServiceBrokers found", func(t *testing.T) {
		// when
		list, _ := cli.listClusterServiceBroker(metav1.ListOptions{})
		// then
		require.Nil(t, list.Items)
	})
}

func TestServiceCatalogClient_ListClusterServiceClass(t *testing.T) {
	//given
	fakeClient := scfake.NewSimpleClientset(newTestCR()...)
	cli := &serviceCatalogClient{client: fakeClient}

	t.Run("should list cluster service classes successfully", func(t *testing.T) {
		// when
		list, err := cli.listClusterServiceClass(metav1.ListOptions{})

		// then
		require.NoError(t, err)
		require.NotNil(t, list.Items)
	})

	//given
	fakeClient = scfake.NewSimpleClientset()
	cli = &serviceCatalogClient{client: fakeClient}

	t.Run("should return nil if no ClusterServiceClasses found", func(t *testing.T) {
		// when
		list, _ := cli.listClusterServiceClass(metav1.ListOptions{})
		// then
		require.Nil(t, list.Items)
	})
}

func TestServiceCatalogClient_ListServiceInstance(t *testing.T) {
	//given
	fakeClient := scfake.NewSimpleClientset(newTestCR()...)
	cli := &serviceCatalogClient{client: fakeClient}

	t.Run("should list service instances successfully", func(t *testing.T) {
		// when
		list, err := cli.listServiceInstance(metav1.ListOptions{})

		// then
		require.NoError(t, err)
		require.NotNil(t, list.Items)
	})

	//given
	fakeClient = scfake.NewSimpleClientset()
	cli = &serviceCatalogClient{client: fakeClient}

	t.Run("should return nil if no ServiceInstances found", func(t *testing.T) {
		// when
		list, _ := cli.listServiceInstance(metav1.ListOptions{})
		// then
		require.Nil(t, list.Items)
	})
}

func TestServiceCatalogClient_FilterCsbWithUrlPrefix(t *testing.T) {
	// given
	fakeClient := scfake.NewSimpleClientset(newTestCR()...)
	cli := &serviceCatalogClient{client: fakeClient}

	brokerList := fixClusterServiceBrokerList()
	brokerUrlPrefix := "https://service-manager."
	expectedSpecUrl := "https://service-manager.katagida.dev"

	t.Run("should return only ClusterServiceBrokers with matching url prefix", func(t *testing.T) {
		// when
		resultList := cli.filterCsbWithUrlPrefix(brokerList, brokerUrlPrefix)

		// then
		assert.Len(t, resultList, 2)
		for _, csb := range resultList {
			assert.EqualValues(t, expectedSpecUrl, csb.Spec.URL)
		}

	})
}

func TestServiceCatalogClient_GetClusterServiceClassesForBrokers(t *testing.T) {
	// given
	fakeClient := scfake.NewSimpleClientset(newTestCR()...)
	cli := &serviceCatalogClient{client: fakeClient}
	expectedBrokenNameSHA := "f17be81c5d87f618a16cfa4e7196494de37016490cd869d740181e2f"

	filteredBrokers := cli.filterCsbWithUrlPrefix(fixClusterServiceBrokerList(), "https://service-manager.")

	t.Run("should list ClusterServiceClasses for ClusterServiceBroker when provided", func(t *testing.T) {
		// when
		resultList, err := cli.getClusterServiceClassesForBrokers(filteredBrokers)

		// then
		require.NoError(t, err)
		assert.Len(t, resultList, 1)
		assert.EqualValues(t, resultList[0].Labels[ClusterServiceBrokerNameLabel], expectedBrokenNameSHA)
	})

	// given
	fakeClient = scfake.NewSimpleClientset()
	cli = &serviceCatalogClient{client: fakeClient}

	t.Run("should return nil if given ClusterServiceBroker do not provide ClusterServiceClass", func(t *testing.T) {
		// when
		resultList, _ := cli.getClusterServiceClassesForBrokers([]v1beta1.ClusterServiceBroker{
			{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1beta1.ClusterServiceBrokerSpec{},
				Status:     v1beta1.ClusterServiceBrokerStatus{},
			},
		})

		// then
		assert.Nil(t, resultList)
	})

}

func TestServiceCatalogClient_GetServiceInstancesForClusterServiceClasses(t *testing.T) {
	// given
	fakeClient := scfake.NewSimpleClientset(newTestCR()...)
	cli := &serviceCatalogClient{client: fakeClient}

	clusterServiceClassList := fixClusterServiceClassList().Items
	expectedLabelValue := "0a3439490f8ad9b70c36c184b1110c044920a99f1b51ce9c3984ae7c"

	t.Run("should list ServiceInstances for ClusterServiceClass", func(t *testing.T) {
		// when
		resultList, err := cli.getServiceInstancesForClusterServiceClasses(clusterServiceClassList)

		// then
		require.NoError(t, err)
		assert.Len(t, resultList, 2)
		for _, serviceInstance := range resultList {
			assert.EqualValues(t, serviceInstance.Labels[ClusterServiceClassRefNameLabel], expectedLabelValue)
		}
	})

	// given
	fakeClient = scfake.NewSimpleClientset()
	cli = &serviceCatalogClient{client: fakeClient}

	t.Run("should return nil if no ServiceInstances found", func(t *testing.T) {
		// when
		resultList, _ := cli.getServiceInstancesForClusterServiceClasses([]v1beta1.ClusterServiceClass{
			{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       v1beta1.ClusterServiceClassSpec{},
				Status:     v1beta1.ClusterServiceClassStatus{},
			},
		})

		// then
		assert.Nil(t, resultList)
	})
}

func fixClusterServiceBrokerList() *v1beta1.ClusterServiceBrokerList {
	return &v1beta1.ClusterServiceBrokerList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []v1beta1.ClusterServiceBroker{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterBrokerNameProvidingClusterServiceClasses,
					Namespace: "compass-system",
				},
				Spec: v1beta1.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
						URL: "https://service-manager.katagida.dev",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ClusterBrokerNameNotProvidingClusterServiceClasses,
					Namespace: "compass-system",
				},
				Spec: v1beta1.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
						URL: "https://service-manager.katagida.dev",
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hb-other-broker",
					Namespace: "kyma-system",
				},
				Spec: v1beta1.ClusterServiceBrokerSpec{
					CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
						URL: "https://other-broker.katagida.dev",
					},
				},
			},
		},
	}

}

func fixClusterServiceClassList() *v1beta1.ClusterServiceClassList {
	return &v1beta1.ClusterServiceClassList{
		TypeMeta: metav1.TypeMeta{},
		ListMeta: metav1.ListMeta{},
		Items: []v1beta1.ClusterServiceClass{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "saas-fake-service-class",
					Labels: map[string]string{ClusterServiceBrokerNameLabel: fixGenerateSHA(ClusterBrokerNameProvidingClusterServiceClasses)},
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					ClusterServiceBrokerName: ClusterBrokerNameProvidingClusterServiceClasses,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   ClusterServiceClassFixedName,
					Labels: map[string]string{ClusterServiceBrokerNameLabel: fixGenerateSHA(ClusterBrokerNameNotProvidingClusterServiceClasses)},
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					ClusterServiceBrokerName: ClusterBrokerNameNotProvidingClusterServiceClasses,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "another-fake-service-class",
					Labels: map[string]string{ClusterServiceBrokerNameLabel: fixGenerateSHA("hb-other-broker")},
				},
				Spec: v1beta1.ClusterServiceClassSpec{
					ClusterServiceBrokerName: "hb-other-broker",
				},
			},
		},
	}
}

func fixGenerateSHA(resourceName string) string {
	return util.GenerateSHA(resourceName)
}

func newTestCR() []runtime.Object {
	return []runtime.Object{
		&v1beta1.ClusterServiceBroker{},
		&v1beta1.ClusterServiceBroker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ClusterBrokerNameProvidingClusterServiceClasses,
				Namespace: "compass-system",
			},
			Spec: v1beta1.ClusterServiceBrokerSpec{
				CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
					URL: "https://service-manager.katagida.dev",
				},
			},
		},
		&v1beta1.ClusterServiceBroker{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hb-other-broker",
				Namespace: "kyma-system",
			},
			Spec: v1beta1.ClusterServiceBrokerSpec{
				CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
					URL: "https://other-broker.katagida.dev",
				},
			},
		},
		&v1beta1.ClusterServiceClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:   ClusterServiceClassFixedName,
				Labels: map[string]string{ClusterServiceBrokerNameLabel: fixGenerateSHA(ClusterBrokerNameProvidingClusterServiceClasses)},
			},
			Spec: v1beta1.ClusterServiceClassSpec{
				ClusterServiceBrokerName: ClusterBrokerNameProvidingClusterServiceClasses,
			},
		},
		&v1beta1.ClusterServiceClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "another-fake-service-class",
				Labels: map[string]string{ClusterServiceBrokerNameLabel: fixGenerateSHA("hb-other-broker")},
			},
			Spec: v1beta1.ClusterServiceClassSpec{
				ClusterServiceBrokerName: "hb-other-broker",
			},
		},
		&v1beta1.ClusterServiceClass{},
		&v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fake-instance-one",
				Namespace: "cstmr-ns",
				Labels: map[string]string{
					ClusterServiceClassRefNameLabel: fixGenerateSHA(ClusterServiceClassFixedName),
				},
			},
		},
		&v1beta1.ServiceInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fake-instance-two",
				Namespace: "cstmr-ns",
				Labels: map[string]string{
					ClusterServiceClassRefNameLabel: fixGenerateSHA(ClusterServiceClassFixedName),
				},
			},
		},
	}
}
