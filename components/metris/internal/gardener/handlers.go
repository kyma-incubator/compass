package gardener

import (
	"fmt"
	"strings"

	gcorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/metris/internal/metrics"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

func (c *Controller) secretUpdateFunc(oldObj, newObj interface{}) {
	newSecret, ok := newObj.(*corev1.Secret)
	if !ok {
		c.logger.Error("error decoding secret, invalid type")
		return
	}

	oldSecret, ok := oldObj.(*corev1.Secret)
	if !ok {
		c.logger.Error("error decoding secret, invalid type")
		return
	}

	if newSecret.ResourceVersion == oldSecret.ResourceVersion {
		// Periodic resync will send update events for all known Secrets, so if the
		// resource version did not change we skip
		return
	}

	// check all shoots with that secret and update
	fselector := fields.SelectorFromSet(fields.Set{fieldSecretBindingName: newSecret.Name}).String()

	shootlist, err := c.gclientset.CoreV1beta1().Shoots(newSecret.Namespace).List(metav1.ListOptions{FieldSelector: fselector})
	if err != nil {
		c.logger.Errorf("error retrieving shoot: %s", err)
		return
	}

	if len(shootlist.Items) == 0 {
		c.logger.Debugf("could not find shoots using secret '%s': skipping", newSecret.Name)
		return
	}

	for i := range shootlist.Items {
		shoot := shootlist.Items[i]
		c.logger.Debugf("secret '%s' has been modified, updating provider client config for shoot '%s'", newSecret.Name, shoot.Name)
		c.shootAddHandlerFunc(&shoot)
	}
}

func (c *Controller) shootDeleteHandlerFunc(obj interface{}) {
	var (
		shoot *gcorev1beta1.Shoot
		ok    bool
	)

	if shoot, ok = obj.(*gcorev1beta1.Shoot); !ok {
		// try to recover deleted obj
		delobj, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			c.logger.Error("error decoding object, invalid type")
			return
		}

		shoot, ok = delobj.Obj.(*gcorev1beta1.Shoot)
		if !ok {
			c.logger.Error("error decoding deleted object, invalid type")
			return
		}
	}

	shootname := fmt.Sprintf("%s--%s", shoot.Namespace, shoot.Name)

	c.logger.Warnf("shoot '%s' is being remove from the provider accounts", shootname)

	c.accountsChan <- &Account{Name: shootname, TechnicalID: shoot.Status.TechnicalID}
}

func (c *Controller) shootAddHandlerFunc(obj interface{}) {
	shoot, ok := obj.(*gcorev1beta1.Shoot)
	if !ok {
		c.logger.Errorf("error decoding shoot object")
		return
	}

	pname := shoot.Spec.CloudProfileName
	if pname == "az" {
		pname = "azure"
	}

	if strings.EqualFold(c.providertype, pname) {
		if shoot.Status.TechnicalID == "" {
			c.logger.Warnf("could not find technical id in shoot '%s', retrying later", shoot.Name)
			c.shootQueue[shoot.Name] = true

			metrics.ClusterSyncFailureVec.WithLabelValues("no_technicalid").Inc()

			return
		}

		accountid, ok := shoot.GetLabels()[labelAccountID]
		if !ok || accountid == "" {
			c.logger.Warnf("could not find label '%s' in Shoot '%s', skipping", labelAccountID, shoot.Name)
			metrics.ClusterSyncFailureVec.WithLabelValues("no_accountid").Inc()

			return
		}

		subaccountid, ok := shoot.GetLabels()[labelSubAccountID]
		if !ok || subaccountid == "" {
			c.logger.Warnf("could not find label '%s' in Shoot '%s', skipping", labelSubAccountID, shoot.Name)
			metrics.ClusterSyncFailureVec.WithLabelValues("no_subaccountid").Inc()

			return
		}

		secret, err := c.kclientset.CoreV1().Secrets(shoot.Namespace).Get(shoot.Spec.SecretBindingName, metav1.GetOptions{})
		if err != nil {
			c.logger.Errorf("error getting shoot secret: %s", err)
			metrics.ClusterSyncFailureVec.WithLabelValues("no_secret").Inc()

			return
		}

		tenantName, ok := secret.GetLabels()[labelTenantName]
		if !ok {
			c.logger.Warnf("could not find label '%s' in secret '%s'", labelTenantName, secret.Name)
		}

		shootName := fmt.Sprintf("%s--%s", shoot.Namespace, shoot.Name)
		shootTechnicalID := shoot.Status.TechnicalID

		c.logger.With("account", accountid).
			With("subaccount", subaccountid).
			With("shoot", shootTechnicalID).Info("New cluster found")

		c.accountsChan <- &Account{
			Name:           shootName,
			ProviderType:   c.providertype,
			AccountID:      accountid,
			SubAccountID:   subaccountid,
			TechnicalID:    shootTechnicalID,
			TenantName:     tenantName,
			CredentialName: secret.Name,
			CredentialData: secret.Data,
		}
	}
}

func (c *Controller) shootUpdateFunc(oldObj, newObj interface{}) {
	newShoot, ok := newObj.(*gcorev1beta1.Shoot)
	if !ok {
		c.logger.Error("error decoding shoot, invalid type")
		return
	}

	if _, ok := c.shootQueue[newShoot.Name]; !ok {
		// shoot was not queued, ignoring update
		return
	}

	// removing shoot from queue
	delete(c.shootQueue, newShoot.Name)

	c.logger.Warnf("shoot '%s' was queued, retrying it", newShoot.Name)

	c.shootAddHandlerFunc(newShoot)
}
