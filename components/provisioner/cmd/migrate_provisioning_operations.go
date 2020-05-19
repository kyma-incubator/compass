package main

import (
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
)

func migrateOperationsInShootProvisioningStage(session dbsession.WriteSession) error {

	// Clean up this code when not needed. TODO: add issue reference here
	err := retry.Do(func() error {
		return session.FixShootProvisioningStage(fmt.Sprintf("Operation in progress. Stage %s", model.WaitingForClusterDomain), model.WaitingForClusterDomain, time.Now())
	}, retry.Attempts(5))

	return err
}
