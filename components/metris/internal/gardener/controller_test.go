package gardener

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	gcorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gfake "github.com/gardener/gardener/pkg/client/core/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
)

var (
	defaultLogger = zap.NewNop().Sugar()

	defaultSecret = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-secret",
			ResourceVersion: "1.1.1",
			Labels: map[string]string{
				labelHyperscalerType: "azure",
				labelTenantName:      "tenantname-test",
			},
			Annotations: map[string]string{},
		},
		Data: map[string][]byte{
			"clientid":     []byte("fakeclientidhere"),
			"clientsecret": []byte("fakeclientsecrethere"),
		},
	}

	defaultShoot = &gcorev1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-shoot",
			Labels: map[string]string{
				labelAccountID:    "accountid-test",
				labelSubAccountID: "subaccountid-test",
			},
		},
		Spec: gcorev1beta1.ShootSpec{
			CloudProfileName:  "az",
			SecretBindingName: "test-secret",
		},
	}
)

func newFakeClient() *Client {
	return &Client{
		Namespace:           "test-ns",
		GardenerClientset:   gfake.NewSimpleClientset(),
		KubernetesClientset: fake.NewSimpleClientset(),
	}
}

func TestController(t *testing.T) {
	asserts := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	accountsChan := make(chan *Account, 1)

	ctrl, err := NewController(newFakeClient(), "azure", accountsChan, defaultLogger)
	if err != nil {
		t.Errorf("NewController() error = %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go ctrl.Run(ctx, &wg)

	// let the informer sync
	time.Sleep(time.Second)

	// Inject initial secret and shoot into the fake client.
	if _, err = ctrl.kclientset.CoreV1().Secrets("test-ns").Create(defaultSecret); err != nil {
		t.Fatalf("error adding secret: %v", err)
	}

	if _, err = ctrl.gclientset.CoreV1beta1().Shoots("test-ns").Create(defaultShoot); err != nil {
		t.Fatalf("error adding shoot: %v", err)
	}

	t.Run("new account", func(t *testing.T) {
		select {
		case <-accountsChan:
		case <-time.After(2 * time.Second):
			t.Error("timed out, did not get the added shoot")
		}
	})

	t.Run("updating shoot secret", func(t *testing.T) {
		newsecret := defaultSecret.DeepCopy()
		newsecret.ObjectMeta.ResourceVersion = "1.1.2"
		newsecret.Data["clientid"] = []byte("new-clientid")

		if _, err = ctrl.kclientset.CoreV1().Secrets("test-ns").Update(newsecret); err != nil {
			t.Fatalf("error updating secret: %v", err)
		}

		select {
		case acc := <-accountsChan:
			asserts.Equalf(acc.CredentialData, newsecret.Data, "account secret should be %s, got %s", string(newsecret.Data["clientid"]), string(acc.CredentialData["clientid"]))
		case <-time.After(2 * time.Second):
			t.Error("timed out, did not get the update")
		}
	})

	t.Run("removing shoot", func(t *testing.T) {
		var gracePeriod int64 = 0
		if err = ctrl.gclientset.CoreV1beta1().Shoots("test-ns").Delete("test-shoot", &metav1.DeleteOptions{GracePeriodSeconds: &gracePeriod}); err != nil {
			t.Fatalf("error adding shoot: %v", err)
		}

		select {
		case acc := <-accountsChan:
			asserts.Emptyf(acc.CredentialData, "account should have been flag for delete: %+v", acc.CredentialData)
		case <-time.After(2 * time.Second):
			t.Error("timed out, did not get the update")
		}
	})

	t.Run("updating secret with no shoot", func(t *testing.T) {
		newsecret := defaultSecret.DeepCopy()
		newsecret.ObjectMeta.ResourceVersion = "1.1.3"
		newsecret.Data["clientid"] = []byte("new-new-clientid")

		if _, err = ctrl.kclientset.CoreV1().Secrets("test-ns").Update(newsecret); err != nil {
			t.Fatalf("error updating secret: %v", err)
		}

		asserts.Never(func() bool {
			<-accountsChan
			return true
		}, 2*time.Second, 2*time.Second, "should not get account update")
	})

	t.Run("adding shoot with missing account label", func(t *testing.T) {
		newshoot := defaultShoot.DeepCopy()
		newshoot.ObjectMeta.Name = "new-test-shoot-account"
		delete(newshoot.ObjectMeta.Labels, labelAccountID)

		if _, err = ctrl.gclientset.CoreV1beta1().Shoots("test-ns").Create(newshoot); err != nil {
			t.Fatalf("error adding shoot: %v", err)
		}

		asserts.Never(func() bool {
			<-accountsChan
			return true
		}, 2*time.Second, 2*time.Second, "should not get account update")
	})

	t.Run("adding shoot with missing subaccount label", func(t *testing.T) {
		newshoot := defaultShoot.DeepCopy()
		newshoot.ObjectMeta.Name = "new-test-shoot-subaccount"
		delete(newshoot.ObjectMeta.Labels, labelSubAccountID)

		if _, err = ctrl.gclientset.CoreV1beta1().Shoots("test-ns").Create(newshoot); err != nil {
			t.Fatalf("error adding shoot: %v", err)
		}

		asserts.Never(func() bool {
			<-accountsChan
			return true
		}, 2*time.Second, 2*time.Second, "should not get account update")
	})

	t.Run("adding shoot with missing secret", func(t *testing.T) {
		newshoot := defaultShoot.DeepCopy()
		newshoot.ObjectMeta.Name = "new-test-shoot-no-secret"
		newshoot.Spec.SecretBindingName = "missing-secret"

		if _, err = ctrl.gclientset.CoreV1beta1().Shoots("test-ns").Create(newshoot); err != nil {
			t.Fatalf("error adding shoot: %v", err)
		}

		asserts.Never(func() bool {
			<-accountsChan
			return true
		}, 2*time.Second, 2*time.Second, "should not get account update")
	})

	cancel()
}
