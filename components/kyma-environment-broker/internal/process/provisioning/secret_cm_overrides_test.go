package provisioning

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestOverridesFromSecretsAndConfigStep_Run(t *testing.T) {
	// Given
	sch := runtime.NewScheme()
	require.NoError(t, coreV1.AddToScheme(sch))
	client := fake.NewFakeClientWithScheme(sch, fixResources()...)

	memoryStorage := storage.NewMemoryStorage()

	inputCreatorMock := &automock.ProvisionInputCreator{}
	defer inputCreatorMock.AssertExpectations(t)
	inputCreatorMock.On("AppendOverrides", "core", []*gqlschema.ConfigEntryInput{
		{
			Key:    "test1",
			Value:  "test1abc",
			Secret: ptr.Bool(true),
		},
		{
			Key:   "test5",
			Value: "test5abc",
		},
	}).Return(nil).Once()
	inputCreatorMock.On("AppendOverrides", "helm", []*gqlschema.ConfigEntryInput{
		{
			Key:    "test3",
			Value:  "test3abc",
			Secret: ptr.Bool(true),
		},
	}).Return(nil).Once()
	inputCreatorMock.On("AppendOverrides", "servicecatalog", []*gqlschema.ConfigEntryInput{
		{
			Key:   "test7",
			Value: "test7abc",
		},
	}).Return(nil).Once()
	inputCreatorMock.On("AppendGlobalOverrides", []*gqlschema.ConfigEntryInput{
		{
			Key:    "test4",
			Value:  "test4abc",
			Secret: ptr.Bool(true),
		},
		{
			Key:   "test8",
			Value: "test8abc",
		},
	}).Return(nil).Once()
	operation := internal.ProvisioningOperation{
		InputCreator: inputCreatorMock,
	}

	step := NewOverridesFromSecretsAndConfigStep(context.TODO(), client, memoryStorage.Operations())

	// When
	operation, repeat, err := step.Run(operation, logrus.New())

	// Then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
}

func fixResources() []runtime.Object {
	var resources []runtime.Object

	resources = append(resources, &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "secret#1",
			Namespace: namespace,
			Labels:    map[string]string{"provisioning-runtime-override": "true", "component": "core"},
		},
		Data: map[string][]byte{"test1": []byte("test1abc")},
	})
	resources = append(resources, &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "secret#2",
			Namespace: namespace,
			Labels:    map[string]string{"component": "core"},
		},
		Data: map[string][]byte{"test2": []byte("test2abc")},
	})
	resources = append(resources, &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "secret#3",
			Namespace: namespace,
			Labels:    map[string]string{"provisioning-runtime-override": "true", "component": "helm"},
		},
		Data: map[string][]byte{"test3": []byte("test3abc")},
	})
	resources = append(resources, &coreV1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "secret#4",
			Namespace: namespace,
			Labels:    map[string]string{"provisioning-runtime-override": "true"},
		},
		Data: map[string][]byte{"test4": []byte("test4abc")},
	})
	resources = append(resources, &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "configmap#1",
			Namespace: namespace,
			Labels:    map[string]string{"provisioning-runtime-override": "true", "component": "core"},
		},
		Data: map[string]string{"test5": "test5abc"},
	})
	resources = append(resources, &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "configmap#2",
			Namespace: "default",
			Labels:    map[string]string{"provisioning-runtime-override": "true", "component": "helm"},
		},
		Data: map[string]string{"test6": "test6abc"},
	})
	resources = append(resources, &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "configmap#3",
			Namespace: namespace,
			Labels:    map[string]string{"provisioning-runtime-override": "true", "component": "servicecatalog"},
		},
		Data: map[string]string{"test7": "test7abc"},
	})
	resources = append(resources, &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "configmap#4",
			Namespace: namespace,
			Labels:    map[string]string{"provisioning-runtime-override": "true"},
		},
		Data: map[string]string{"test8": "test8abc"},
	})

	return resources
}
