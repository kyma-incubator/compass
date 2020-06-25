package environmentscleanup

import (
	"testing"
	"time"

	"github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	mocks "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/environmentscleanup/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	fixInstanceID      = "72b83910-ac12-4dcb-b91d-960cca2b36abx"
	fixRuntimeID       = "2498c8ee-803a-43c2-8194-6d6dd0354c30"
	fixOperationID     = "17f3ddba-1132-466d-a3c5-920f544d7ea6"
	maxShootAge        = 24 * time.Hour
	shootLabelSelector = "owner.do-not-delete!=true"
)

func TestService_PerformCleanup(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// given
		gcMock := &mocks.GardenerClient{}
		gcMock.On("List", mock.AnythingOfType("v1.ListOptions")).Return(fixShootList(), nil)
		bcMock := &mocks.BrokerClient{}
		bcMock.On("Deprovision", mock.AnythingOfType("DeprovisionDetails")).Return(fixOperationID, nil)

		memoryStorage := storage.NewMemoryStorage()
		memoryStorage.Instances().Insert(internal.Instance{
			InstanceID: fixInstanceID,
			RuntimeID:  fixRuntimeID,
		})
		memoryStorage.Instances().Insert(internal.Instance{
			InstanceID: "second-instance",
			RuntimeID:  "some-runtime-id",
		})

		svc := NewService(gcMock, bcMock, memoryStorage.Instances(), maxShootAge, shootLabelSelector)

		// when
		err := svc.PerformCleanup()

		// then
		assert.NoError(t, err)
	})

	t.Run("should fail when unable to fetch shoots from gardener", func(t *testing.T) {
		// given
		gcMock := &mocks.GardenerClient{}
		gcMock.On("List", mock.AnythingOfType("v1.ListOptions")).Return(&v1beta1.ShootList{}, errors.New("failed to reach gardener"))
		bcMock := &mocks.BrokerClient{}
		memoryStorage := storage.NewMemoryStorage()

		svc := NewService(gcMock, bcMock, memoryStorage.Instances(), maxShootAge, shootLabelSelector)

		// when
		err := svc.PerformCleanup()

		// then
		assert.Error(t, err)
	})

	t.Run("should return error when unable to find instance in db", func(t *testing.T) {
		// given
		gcMock := &mocks.GardenerClient{}
		gcMock.On("List", mock.AnythingOfType("v1.ListOptions")).Return(fixShootList(), nil)
		bcMock := &mocks.BrokerClient{}
		bcMock.On("Deprovision", mock.AnythingOfType("DeprovisionDetails")).Return(fixOperationID, nil)

		memoryStorage := storage.NewMemoryStorage()
		memoryStorage.Instances().Insert(internal.Instance{
			InstanceID: "some-instance-id",
			RuntimeID:  "not-matching-id",
		})

		svc := NewService(gcMock, bcMock, memoryStorage.Instances(), maxShootAge, shootLabelSelector)

		// when
		err := svc.PerformCleanup()

		// then
		assert.Error(t, err)
	})

	t.Run("should return error on deprovision call failure", func(t *testing.T) {
		// given
		gcMock := &mocks.GardenerClient{}
		gcMock.On("List", mock.AnythingOfType("v1.ListOptions")).Return(fixShootList(), nil)
		bcMock := &mocks.BrokerClient{}
		bcMock.On("Deprovision", mock.AnythingOfType("DeprovisionDetails")).Return("", errors.New("failed to deprovision instance"))

		memoryStorage := storage.NewMemoryStorage()
		memoryStorage.Instances().Insert(internal.Instance{
			InstanceID: fixInstanceID,
			RuntimeID:  fixRuntimeID,
		})

		svc := NewService(gcMock, bcMock, memoryStorage.Instances(), maxShootAge, shootLabelSelector)

		// when
		err := svc.PerformCleanup()

		// then
		assert.Error(t, err)
	})

	t.Run("should return error when shoot has no runtime id", func(t *testing.T) {
		// given
		gcMock := &mocks.GardenerClient{}
		creationTime, parseErr := time.Parse(time.RFC3339, "2020-01-02T10:00:00-05:00")
		require.NoError(t, parseErr)
		gcMock.On("List", mock.AnythingOfType("v1.ListOptions")).Return(&v1beta1.ShootList{
			TypeMeta: v1.TypeMeta{},
			ListMeta: v1.ListMeta{},
			Items: []v1beta1.Shoot{
				{
					TypeMeta: v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{
						Name:              "az-1234",
						CreationTimestamp: v1.Time{Time: creationTime},
						Labels:            map[string]string{"should-be-deleted": "true"},
						Annotations:       map[string]string{"created-by": "not-provisioner"},
						ClusterName:       "cluster-one",
					},
					Spec: v1beta1.ShootSpec{
						CloudProfileName: "az",
					},
				}},
		}, nil)
		bcMock := &mocks.BrokerClient{}
		bcMock.On("Deprovision", mock.AnythingOfType("DeprovisionDetails")).Return("", errors.New("failed to deprovision instance"))

		memoryStorage := storage.NewMemoryStorage()
		memoryStorage.Instances().Insert(internal.Instance{
			InstanceID: fixInstanceID,
			RuntimeID:  fixRuntimeID,
		})

		svc := NewService(gcMock, bcMock, memoryStorage.Instances(), maxShootAge, shootLabelSelector)

		// when
		err := svc.PerformCleanup()

		// then
		assert.Error(t, err)
	})
}

func fixShootList() *v1beta1.ShootList {
	return &v1beta1.ShootList{
		TypeMeta: v1.TypeMeta{},
		ListMeta: v1.ListMeta{},
		Items:    fixShootListItems(),
	}
}

func fixShootListItems() []v1beta1.Shoot {
	creationTime, _ := time.Parse(time.RFC3339, "2020-01-02T10:00:00-05:00")

	return []v1beta1.Shoot{
		{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name:              "az-1234",
				CreationTimestamp: v1.Time{Time: creationTime},
				Labels:            map[string]string{"should-be-deleted": "true"},
				Annotations:       map[string]string{shootAnnotationRuntimeId: "some-runtime-id"},
				ClusterName:       "cluster-one",
			},
			Spec: v1beta1.ShootSpec{
				CloudProfileName: "az",
			},
		},
		{
			TypeMeta: v1.TypeMeta{},
			ObjectMeta: v1.ObjectMeta{
				Name:              "gcp-1234",
				CreationTimestamp: v1.Time{Time: creationTime},
				Labels:            map[string]string{"some-label": "some-value"},
				Annotations:       map[string]string{shootAnnotationRuntimeId: fixRuntimeID},
				ClusterName:       "cluster-two",
			},
			Spec: v1beta1.ShootSpec{
				CloudProfileName: "gcp",
			},
		},
	}
}
