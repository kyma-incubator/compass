package main

import (
	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

func initPersistence(connectionString, schemaPath string) (persistence.RuntimeService, persistence.OperationService, error) {
	_, err := database.InitializeDatabase(connectionString, schemaPath)
	if err != nil {
		return nil, nil, err
	}

	connection, err := dbr.Open(connectionString, "postgres", nil)
	dbSessionFactory := dbsession.NewDBSessionFactory(connection)

	runtimeService := persistence.NewRuntimeService(dbSessionFactory)
	operationService := persistence.NewOperationService(dbSessionFactory)

	return runtimeService, operationService, nil
}

func initProvisioningService(runtimeService persistence.RuntimeService, operationService persistence.OperationService, secrets v1.SecretInterface) provisioning.ProvisioningService {
	hydroformClient := hydroform.NewHydroformClient(secrets)

	return provisioning.NewProvisioningService(operationService, runtimeService, hydroformClient)
}
