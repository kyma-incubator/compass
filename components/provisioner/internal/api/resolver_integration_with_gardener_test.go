package api

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/api/middlewares"
	mocks2 "github.com/kyma-incubator/compass/components/provisioner/internal/runtime/clientbuilder/mocks"
	"k8s.io/client-go/kubernetes/fake"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	directormock "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/gardener"
	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/testutils"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	runtimeConfigrtr "github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/install/installation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var testEnv *envtest.Environment
var cfg *rest.Config
var mgr ctrl.Manager

const (
	namespace  = "default"
	timeout    = 10 * time.Second
	syncPeriod = 5 * time.Second
	waitPeriod = syncPeriod + 3*time.Second

	mockedKubeconfig = `apiVersion: v1
clusters:
- cluster:
    server: https://192.168.64.4:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURBRENDQWVpZ0F3SUJBZ0lCQWpBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwdGFXNXAKYTNWaVpVTkJNQjRYRFRFNU1URXhOekE0TXpBek1sb1hEVEl3TVRFeE56QTRNekF6TWxvd01URVhNQlVHQTFVRQpDaE1PYzNsemRHVnRPbTFoYzNSbGNuTXhGakFVQmdOVkJBTVREVzFwYm1scmRXSmxMWFZ6WlhJd2dnRWlNQTBHCkNTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDNmY2SjZneElvL2cyMHArNWhybklUaUd5SDh0VW0KWGl1OElaK09UKyt0amd1OXRneXFnbnNsL0dDT1Q3TFo4ejdOVCttTEdKL2RLRFdBV3dvbE5WTDhxMzJIQlpyNwpDaU5hK3BBcWtYR0MzNlQ2NEQyRjl4TEtpVVpuQUVNaFhWOW1oeWVCempscTh1NnBjT1NrY3lJWHRtdU9UQUVXCmErWlp5UlhOY3BoYjJ0NXFUcWZoSDhDNUVDNUIrSm4rS0tXQ2Y1Nm5KZGJQaWduRXh4SFlaMm9TUEc1aXpkbkcKZDRad2d0dTA3NGttaFNtNXQzbjgyNmovK29tL25VeWdBQ24yNmR1K21aZzRPcWdjbUMrdnBYdUEyRm52bk5LLwo5NWErNEI3cGtNTER1bHlmUTMxcjlFcStwdHBkNUR1WWpldVpjS1Bxd3ZVcFUzWVFTRUxVUzBrUkFnTUJBQUdqClB6QTlNQTRHQTFVZER3RUIvd1FFQXdJRm9EQWRCZ05WSFNVRUZqQVVCZ2dyQmdFRkJRY0RBUVlJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBQ3JnbExWemhmemZ2aFNvUgowdWNpNndBZDF6LzA3bW52MDRUNmQyTkpjRG80Uzgwa0o4VUJtRzdmZE5qMlJEaWRFbHRKRU1kdDZGa1E1TklOCk84L1hJdENiU0ZWYzRWQ1NNSUdPcnNFOXJDajVwb24vN3JxV3dCbllqYStlbUVYOVpJelEvekJGU3JhcWhud3AKTkc1SmN6bUg5ODRWQUhGZEMvZWU0Z2szTnVoV25rMTZZLzNDTTFsRkxlVC9Cbmk2K1M1UFZoQ0x3VEdmdEpTZgorMERzbzVXVnFud2NPd3A3THl2K3h0VGtnVmdSRU5RdTByU2lWL1F2UkNPMy9DWXdwRTVIRFpjalM5N0I4MW0yCmVScVBENnVoRjVsV3h4NXAyeEd1V2JRSkY0WnJzaktLTW1CMnJrUnR5UDVYV2xWZU1mR1VjbFdjc1gxOW91clMKaWpKSTFnPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBdW4raWVvTVNLUDROdEtmdVlhNXlFNGhzaC9MVkpsNHJ2Q0dmamsvdnJZNEx2YllNCnFvSjdKZnhnamsreTJmTSt6VS9waXhpZjNTZzFnRnNLSlRWUy9LdDlod1dhK3dvald2cVFLcEZ4Z3Qrayt1QTkKaGZjU3lvbEdad0JESVYxZlpvY25nYzQ1YXZMdXFYRGtwSE1pRjdacmprd0JGbXZtV2NrVnpYS1lXOXJlYWs2bgo0Ui9BdVJBdVFmaVovaWlsZ24rZXB5WFd6NG9KeE1jUjJHZHFFanh1WXMzWnhuZUdjSUxidE8rSkpvVXB1YmQ1Ci9OdW8vL3FKdjUxTW9BQXA5dW5idnBtWU9EcW9ISmd2cjZWN2dOaFo3NXpTdi9lV3Z1QWU2WkRDdzdwY24wTjkKYS9SS3ZxYmFYZVE3bUkzcm1YQ2o2c0wxS1ZOMkVFaEMxRXRKRVFJREFRQUJBb0lCQVFDTEVFa3pXVERkYURNSQpGb0JtVGhHNkJ1d0dvMGZWQ0R0TVdUWUVoQTZRTjI4QjB4RzJ3dnpZNGt1TlVsaG10RDZNRVo1dm5iajJ5OWk1CkVTbUxmU3VZUkxlaFNzaTVrR0cwb1VtR3RGVVQ1WGU3cWlHMkZ2bm9GRnh1eVg5RkRiN3BVTFpnMEVsNE9oVkUKTzI0Q1FlZVdEdXc4ZXVnRXRBaGJ3dG1ERElRWFdPSjcxUEcwTnZKRHIwWGpkcW1aeExwQnEzcTJkZTU2YmNjawpPYzV6dmtJNldrb0o1TXN0WkZpU3pVRDYzN3lIbjh2NGd3cXh0bHFoNWhGLzEwV296VmZqVGdWSG0rc01ZaU9SCmNIZ0dMNUVSbDZtVlBsTTQzNUltYnFnU1R2NFFVVGpzQjRvbVBsTlV5Yksvb3pPSWx3RjNPTkJjVVV6eDQ1cGwKSHVJQlQwZ1JBb0dCQU9SR2lYaVBQejdsay9Bc29tNHkxdzFRK2hWb3Yvd3ovWFZaOVVkdmR6eVJ1d3gwZkQ0QgpZVzlacU1hK0JodnB4TXpsbWxYRHJBMklYTjU3UEM3ZUo3enhHMEVpZFJwN3NjN2VmQUN0eDN4N0d0V2pRWGF2ClJ4R2xDeUZxVG9LY3NEUjBhQ0M0Um15VmhZRTdEY0huLy9oNnNzKys3U2tvRVMzNjhpS1RiYzZQQW9HQkFORW0KTHRtUmZieHIrOE5HczhvdnN2Z3hxTUlxclNnb2NmcjZoUlZnYlU2Z3NFd2pMQUs2ZHdQV0xWQmVuSWJ6bzhodApocmJHU1piRnF0bzhwS1Q1d2NxZlpKSlREQnQxYmhjUGNjWlRmSnFmc0VISXc0QW5JMVdRMlVzdzVPcnZQZWhsCmh0ek95cXdBSGZvWjBUTDlseTRJUHRqbXArdk1DQ2NPTHkwanF6NWZBb0dCQUlNNGpRT3hqSkN5VmdWRkV5WTMKc1dsbE9DMGdadVFxV3JPZnY2Q04wY1FPbmJCK01ZRlBOOXhUZFBLeC96OENkVyszT0syK2FtUHBGRUdNSTc5cApVdnlJdUxzTGZMZDVqVysyY3gvTXhaU29DM2Z0ZmM4azJMeXEzQ2djUFA5VjVQQnlUZjBwRU1xUWRRc2hrRG44CkRDZWhHTExWTk8xb3E5OTdscjhMY3A2L0FvR0FYNE5KZC9CNmRGYjRCYWkvS0lGNkFPQmt5aTlGSG9iQjdyVUQKbTh5S2ZwTGhrQk9yNEo4WkJQYUZnU09ENWhsVDNZOHZLejhJa2tNNUVDc0xvWSt4a1lBVEpNT3FUc3ZrOThFRQoyMlo3Qy80TE55K2hJR0EvUWE5Qm5KWDZwTk9XK1ErTWRFQTN6QzdOZ2M3U2U2L1ZuNThDWEhtUmpCeUVTSm13CnI3T1BXNDhDZ1lBVUVoYzV2VnlERXJxVDBjN3lIaXBQbU1wMmljS1hscXNhdC94YWtobENqUjZPZ2I5aGQvNHIKZm1wUHJmd3hjRmJrV2tDRUhJN01EdDJrZXNEZUhRWkFxN2xEdjVFT2k4ZG1uM0ZPNEJWczhCOWYzdm52MytmZwpyV2E3ZGtyWnFudU12cHhpSWlqOWZEak9XbzdxK3hTSFcxWWdSNGV2Q1p2NGxJU0FZRlViemc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`
)

func TestMain(m *testing.M) {
	err := setupEnv()
	if err != nil {
		logrus.Errorf("Failed to setup test environment: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		err := testEnv.Stop()
		if err != nil {
			logrus.Errorf("error while deleting Compass Connection: %s", err.Error())
		}
	}()

	syncPeriod := syncPeriod

	mgr, err = ctrl.NewManager(cfg, ctrl.Options{SyncPeriod: &syncPeriod, Namespace: namespace})

	if err != nil {
		logrus.Errorf("unable to create shoot controller mgr: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func setupEnv() error {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("testdata")},
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		return errors.Wrap(err, "Failed to start test environment")
	}

	return nil
}

func TestProvisioning_ProvisionRuntimeWithDatabase(t *testing.T) {
	//given
	installationServiceMock := &installationMocks.Service{}
	installationServiceMock.On("InstallKyma", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("model.Release"),
		mock.AnythingOfType("model.Configuration"), mock.AnythingOfType("[]model.KymaComponentConfig")).Return(nil)

	installationServiceMock.On("TriggerInstallation", mock.Anything, mock.AnythingOfType("model.Release"),
		mock.AnythingOfType("model.Configuration"), mock.AnythingOfType("[]model.KymaComponentConfig")).Return(nil)

	installationServiceMock.On("CheckInstallationState", mock.Anything).Return(installation.InstallationState{State: "Installed"}, nil)
	installationServiceMock.On("TriggerUninstall", mock.Anything).Return(nil)

	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)

	cleanupNetwork, err := testutils.EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	containerCleanupFunc, connString, err := testutils.InitTestDBContainer(t, ctx, "postgres_database_2")
	require.NoError(t, err)

	defer containerCleanupFunc()

	connection, err := database.InitializeDatabaseConnection(connString, 5)
	require.NoError(t, err)
	require.NotNil(t, connection)

	defer testutils.CloseDatabase(t, connection)

	err = database.SetupSchema(connection, testutils.SchemaFilePath)
	require.NoError(t, err)

	kymaConfig := fixKymaGraphQLConfigInput()

	clusterConfigurations := getTestClusterConfigurations()
	directorServiceMock := &directormock.DirectorClient{}

	cmClientBuilder := &mocks2.ConfigMapClientBuilder{}
	configMapClient := fake.NewSimpleClientset().CoreV1().ConfigMaps(compassSystemNamespace)
	cmClientBuilder.On("CreateK8SConfigMapClient", mockedKubeconfig, compassSystemNamespace).Return(configMapClient, nil)
	runtimeConfigurator := runtimeConfigrtr.NewRuntimeConfigurator(cmClientBuilder, directorServiceMock)

	shootInterface := newFakeShootsInterface(t, cfg)
	secretsInterface := setupSecretsClient(t, cfg)
	dbsFactory := dbsession.NewFactory(connection)

	controler, err := gardener.NewShootController(namespace, mgr, shootInterface, secretsInterface, installationServiceMock, dbsFactory, timeout, directorServiceMock, runtimeConfigurator)
	require.NoError(t, err)

	go func() {
		err := controler.StartShootController()
		require.NoError(t, err)
	}()

	for _, config := range clusterConfigurations {
		t.Run(config.description, func(t *testing.T) {
			configMapClient.Delete(runtimeConfigrtr.ConfigMapName, &metav1.DeleteOptions{})

			directorServiceMock.Calls = nil
			directorServiceMock.ExpectedCalls = nil

			directorServiceMock.On("CreateRuntime", mock.Anything, mock.Anything).Return(config.runtimeID, nil)
			directorServiceMock.On("DeleteRuntime", mock.Anything, mock.Anything).Return(nil)
			directorServiceMock.On("GetConnectionToken", mock.Anything, mock.Anything).Return(graphql.OneTimeTokenForRuntimeExt{}, nil)

			uuidGenerator := uuid.NewUUIDGenerator()
			provisioner := gardener.NewProvisioner(namespace, shootInterface)

			releaseRepository := release.NewReleaseRepository(connection, uuidGenerator)

			inputConverter := provisioning.NewInputConverter(uuidGenerator, releaseRepository, "Project")
			graphQLConverter := provisioning.NewGraphQLConverter()

			provisioningService := provisioning.NewProvisioningService(inputConverter, graphQLConverter, directorServiceMock, dbsFactory, provisioner, uuidGenerator)

			validator := NewValidator(dbsFactory.NewReadSession())

			resolver := NewResolver(provisioningService, validator)

			err = insertDummyReleaseIfNotExist(releaseRepository, uuidGenerator.New(), kymaVersion)
			require.NoError(t, err)

			fullConfig := gqlschema.ProvisionRuntimeInput{RuntimeInput: config.runtimeInput, ClusterConfig: config.config, Credentials: providerCredentials, KymaConfig: kymaConfig}

			//when
			provisionRuntime, err := resolver.ProvisionRuntime(ctx, fullConfig)

			//then
			require.NoError(t, err)
			require.NotEmpty(t, provisionRuntime)

			//when
			//wait for Shoot to update
			time.Sleep(waitPeriod)

			list, err := shootInterface.List(metav1.ListOptions{})
			require.NoError(t, err)

			shoot := &list.Items[0]

			//then
			assert.Equal(t, config.runtimeID, shoot.Annotations["compass.provisioner.kyma-project.io/runtime-id"])
			assert.Equal(t, *provisionRuntime.ID, shoot.Annotations["compass.provisioner.kyma-project.io/operation-id"])
			assert.Equal(t, "provisioning", shoot.Annotations["compass.provisioner.kyma-project.io/provisioning"])
			assert.Equal(t, "provisioning-in-progress", shoot.Annotations["compass.provisioner.kyma-project.io/provisioning-step"])

			simmulateSuccessfullClusterProvisioning(t, shootInterface, secretsInterface, shoot)

			//when
			//wait for Shoot to update
			time.Sleep(waitPeriod)
			shoot, err = shootInterface.Get(shoot.Name, metav1.GetOptions{})

			//then
			require.NoError(t, err)
			assert.Equal(t, config.runtimeID, shoot.Annotations["compass.provisioner.kyma-project.io/runtime-id"])
			assert.Equal(t, "provisioning-finished", shoot.Annotations["compass.provisioner.kyma-project.io/provisioning-step"])
			assert.Equal(t, "provisioned", shoot.Annotations["compass.provisioner.kyma-project.io/provisioning"])
			assert.Equal(t, "installed", shoot.Annotations["compass.provisioner.kyma-project.io/kyma-installation"])

			//when
			deprovisionRuntimeID, err := resolver.DeprovisionRuntime(ctx, config.runtimeID)

			//then
			require.NoError(t, err)
			require.NotEmpty(t, deprovisionRuntimeID)

			//when
			//wait for Shoot to update
			time.Sleep(waitPeriod)
			shoot, err = shootInterface.Get(shoot.Name, metav1.GetOptions{})

			//then
			require.NoError(t, err)
			assert.Equal(t, config.runtimeID, shoot.Annotations["compass.provisioner.kyma-project.io/runtime-id"])
			assert.Equal(t, "deprovisioning", shoot.Annotations["compass.provisioner.kyma-project.io/provisioning-step"])
			assert.Equal(t, deprovisionRuntimeID, shoot.Annotations["compass.provisioner.kyma-project.io/operation-id"])
			assert.Equal(t, "deprovisioning", shoot.Annotations["compass.provisioner.kyma-project.io/provisioning"])
			assert.Equal(t, "uninstalling", shoot.Annotations["compass.provisioner.kyma-project.io/kyma-installation"])

			//when
			shoot = removeFinalizers(t, shootInterface, shoot)
			time.Sleep(waitPeriod)
			shoot, err = shootInterface.Get(shoot.Name, metav1.GetOptions{})

			//then
			require.Error(t, err)
			require.Empty(t, shoot)
		})
	}
}

func newFakeShootsInterface(t *testing.T, config *rest.Config) gardener_apis.ShootInterface {
	dynamicClient, err := dynamic.NewForConfig(config)
	require.NoError(t, err)

	resourceInterface := dynamicClient.Resource(gardener_types.SchemeGroupVersion.WithResource("shoots"))
	return &fakeShootsInterface{
		client: resourceInterface,
	}
}

type fakeShootsInterface struct {
	client dynamic.ResourceInterface
}

func (f fakeShootsInterface) Create(shoot *gardener_types.Shoot) (*gardener_types.Shoot, error) {
	addTypeMeta(shoot)

	shoot.SetFinalizers([]string{"finalizer"})

	unstructuredShoot, err := toUnstructured(shoot)

	if err != nil {
		return nil, err
	}

	create, err := f.client.Create(unstructuredShoot, metav1.CreateOptions{})

	if err != nil {
		return nil, err
	}

	return fromUnstructured(create)
}

func removeFinalizers(t *testing.T, shootInterface gardener_apis.ShootInterface, shoot *gardener_types.Shoot) *gardener_types.Shoot {
	shoot.SetFinalizers([]string{})

	update, err := shootInterface.Update(shoot)
	require.NoError(t, err)
	return update
}

func (f *fakeShootsInterface) Update(shoot *gardener_types.Shoot) (*gardener_types.Shoot, error) {
	obj, err := toUnstructured(shoot)

	if err != nil {
		return nil, err
	}

	updated, err := f.client.Update(obj, metav1.UpdateOptions{})

	if err != nil {
		return nil, err
	}

	return fromUnstructured(updated)
}

func (f *fakeShootsInterface) UpdateStatus(*gardener_types.Shoot) (*gardener_types.Shoot, error) {
	return nil, nil
}

func (f *fakeShootsInterface) Delete(name string, options *metav1.DeleteOptions) error {
	return f.client.Delete(name, options)
}

func (f *fakeShootsInterface) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return nil
}

func (f *fakeShootsInterface) Get(name string, options metav1.GetOptions) (*gardener_types.Shoot, error) {
	obj, err := f.client.Get(name, options)

	if err != nil {
		return nil, err
	}

	return fromUnstructured(obj)
}
func (f *fakeShootsInterface) List(opts metav1.ListOptions) (*gardener_types.ShootList, error) {
	list, err := f.client.List(opts)

	if err != nil {
		return nil, err
	}

	return listFromUnstructured(list)
}
func (f *fakeShootsInterface) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}
func (f *fakeShootsInterface) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *gardener_types.Shoot, err error) {
	return nil, nil
}

func simmulateSuccessfullClusterProvisioning(t *testing.T, f gardener_apis.ShootInterface, s v1core.SecretInterface, shoot *gardener_types.Shoot) {
	setShootStatusToSuccessfull(t, f, shoot)
	createKubeconfigSecret(t, s, shoot.Name)
}

func setShootStatusToSuccessfull(t *testing.T, f gardener_apis.ShootInterface, shoot *gardener_types.Shoot) {
	shoot.Status.LastOperation = &gardener_types.LastOperation{State: gardener_types.LastOperationStateSucceeded}

	_, err := f.Update(shoot)

	require.NoError(t, err)
}

func createKubeconfigSecret(t *testing.T, s v1core.SecretInterface, shootName string) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.kubeconfig", shootName),
			Namespace: namespace,
		},
		Data: map[string][]byte{"kubeconfig": []byte(mockedKubeconfig)},
	}
	_, err := s.Create(secret)

	require.NoError(t, err)
}

func addTypeMeta(shoot *gardener_types.Shoot) {
	shoot.TypeMeta = metav1.TypeMeta{
		Kind:       "Shoot",
		APIVersion: "core.gardener.cloud/v1beta1",
	}
}

func toUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)

	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: object}, nil
}

func fromUnstructured(object *unstructured.Unstructured) (*gardener_types.Shoot, error) {
	var newShoot gardener_types.Shoot

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(object.Object, &newShoot)

	if err != nil {
		return nil, err
	}

	return &newShoot, err
}

func listFromUnstructured(list *unstructured.UnstructuredList) (*gardener_types.ShootList, error) {
	shootList := &gardener_types.ShootList{
		Items: []gardener_types.Shoot{},
	}

	for _, obj := range list.Items {
		shoot, err := fromUnstructured(&obj)
		if err != nil {
			return &gardener_types.ShootList{}, err
		}
		shootList.Items = append(shootList.Items, *shoot)
	}
	return shootList, nil
}

func setupSecretsClient(t *testing.T, config *rest.Config) v1core.SecretInterface {
	coreClient, err := v1core.NewForConfig(config)
	require.NoError(t, err)

	return coreClient.Secrets(namespace)
}
