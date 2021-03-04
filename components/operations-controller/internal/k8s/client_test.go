/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8s_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/k8s"
	"github.com/stretchr/testify/require"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

const webhookID = "866e6b9c-f03b-442b-a6a5-4b90e21e503a"

func TestClient(t *testing.T) {
	ctx := context.Background()

	testEnv := &envtest.Environment{
		ErrorIfCRDPathMissing: true,
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err)
	defer func() {
		err := testEnv.Stop()
		require.NoError(t, err)
	}()

	err = v1alpha1.AddToScheme(scheme.Scheme)
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})

	client := k8s.NewClient(k8sClient)

	operation := &v1alpha1.Operation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-operation",
			Namespace: "default",
		},
		Spec: v1alpha1.OperationSpec{
			OperationType: "Delete",
			WebhookIDs:    []string{webhookID},
		},
	}

	err = client.Create(ctx, operation)
	require.NoError(t, err)

	namespacedName := types.NamespacedName{Namespace: operation.ObjectMeta.Namespace, Name: operation.ObjectMeta.Name}

	t.Run("Get returns existing operation successfully", func(t *testing.T) {
		actualOperation, err := client.Get(ctx, namespacedName)
		require.NoError(t, err)

		require.Equal(t, operation.ObjectMeta.Name, actualOperation.ObjectMeta.Name)
		require.Equal(t, operation.ObjectMeta.Namespace, actualOperation.ObjectMeta.Namespace)

		require.Equal(t, operation.Spec.OperationType, actualOperation.Spec.OperationType)
		require.Equal(t, operation.Spec.WebhookIDs, actualOperation.Spec.WebhookIDs)
	})

	t.Run("Get returns existing error when trying to fetch non-existant operation", func(t *testing.T) {
		_, err := client.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      "non-existant-operation",
		})

		require.Error(t, err)
		require.True(t, kubeerrors.IsNotFound(err))
	})
}
