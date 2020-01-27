package gardener

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"

	v1core "k8s.io/client-go/kubernetes/typed/core/v1"

	restclient "k8s.io/client-go/rest"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	gardener_apis "github.com/gardener/gardener/pkg/client/garden/clientset/versioned/typed/garden/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"github.com/sirupsen/logrus"

	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"

	ctrl "sigs.k8s.io/controller-runtime"
)

// TODO: refactor
func NewShootController(
	namespace string,
	k8sConfig *restclient.Config,
	shootClient gardener_apis.ShootInterface,
	secretsClient v1core.SecretInterface,
	installationService installation.Service,
	dbsFactory dbsession.Factory,
	installationTimeout time.Duration,
	directorClient director.DirectorClient) (*ShootController, error) {
	defaultSyncPeriod := 3 * time.Minute

	mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{SyncPeriod: &defaultSyncPeriod, Namespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("unable to create shoot controller manager: %w", err)
	}

	err = gardener_types.AddToScheme(mgr.GetScheme())
	if err != nil {
		return nil, fmt.Errorf("failed to add Gardener types to scheme: %s", err.Error())
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&gardener_types.Shoot{}).
		Complete(NewReconciler(mgr, dbsFactory, secretsClient, shootClient, installationService, installationTimeout, directorClient))
	if err != nil {
		return nil, fmt.Errorf("unable to create controller: %w", err)
	}

	return &ShootController{
		namespace:         namespace,
		controllerManager: mgr,
		shootClient:       shootClient,
		dbSessionFactory:  dbsFactory,
		log:               logrus.WithField("Component", "ShootController"),
	}, nil
}

type ShootController struct {
	namespace         string
	controllerManager ctrl.Manager
	dbSessionFactory  dbsession.Factory
	shootClient       gardener_apis.ShootInterface
	log               *logrus.Entry
}

func (sc *ShootController) StartShootController() error {
	session := sc.dbSessionFactory.NewReadSession()

	// List all gardenerClusters in db
	gardenerClusters, dberr := session.ListGardenerClusters()
	if dberr != nil {
		return dberr
	}

	// List all shoots
	shootList, err := sc.shootClient.List(v1.ListOptions{})
	if err != nil {
		return err
	}

	// TODO: this logic can be used to recover from some downtime while provisioning was in progress
	// Create shoots for clusters that exist in database but does not have a corresponding shoot
	for _, c := range gardenerClusters {
		gardenerConfig, ok := c.GardenerConfig()
		if !ok {
			return fmt.Errorf("Cluster %s does not have Gardener configuration", c.ID)
		}

		shoot, found := getShootByName(shootList, gardenerConfig.Name)
		if !found && !c.Deleted {
			sc.log.Infof("Creating shoot: %s", shoot.Name)
			shootTemplate := gardenerConfig.ToShootTemplate(sc.namespace)
			annotate(shootTemplate, runtimeIdAnnotation, c.ID)

			shoot, err := sc.shootClient.Create(shootTemplate)
			if err != nil {
				return fmt.Errorf("error creating shoot: %s", err.Error())
			}
			sc.log.Infof("Shoot %s created", shoot.Name)
		}

		// TODO: here we can ensure that Shoot state is matching this from Database
	}
	//
	//// TODO: here we can delete Shoots that does not exist in database
	//// Delete shoots that does not have the corresponding cluster
	//for _, s := range shootList.Items {
	//	_, found := getGardenerClusterByName(gardenerClusters, s.Name)
	//	if !found {
	//		if s.DeletionTimestamp != nil {
	//			sc.log.Infof("Shoot %s already scheduled for deletion", s.Name)
	//			continue
	//		}
	//		sc.log.Infof("Deleting Shoot %s as it does not exist in Database", s.Name)
	//		AnnotateWithConfirmDeletion(&s)
	//		err = UpdateAndDeleteShoot(sc.shootClient, &s)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}

	// Start Controller
	if err := sc.controllerManager.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error starting shoot controller: %w", err)
	}

	return nil
}

func getGardenerClusterByName(gardenerClusters []model.Cluster, name string) (model.GardenerConfig, bool) {
	for _, c := range gardenerClusters {
		gardenerConfig, ok := c.GardenerConfig()
		if !ok {
			panic("NOT A GARDENER")
		}

		if gardenerConfig.Name == name {
			return gardenerConfig, true
		}
	}

	return model.GardenerConfig{}, false
}

func getShootByName(shootList *gardener_types.ShootList, name string) (gardener_types.Shoot, bool) {
	for _, s := range shootList.Items {
		if s.Name == name {
			return s, true
		}
	}

	return gardener_types.Shoot{}, false
}
