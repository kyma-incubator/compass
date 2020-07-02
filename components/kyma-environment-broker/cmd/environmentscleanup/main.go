package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dlmiddlecote/sqlstats"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/environmentscleanup"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/environmentscleanup/broker"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/gardener"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type config struct {
	MaxAgeHours   time.Duration `envconfig:"default=24h"`
	LabelSelector string        `envconfig:"default=owner.do-not-delete!=true"`
	Gardener      gardener.Config
	Database      storage.Config
	Broker        broker.Config
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(errors.Wrap(err, "while loading cleanup config"))

	clusterCfg, err := gardener.NewGardenerClusterConfig(cfg.Gardener.KubeconfigPath)
	fatalOnError(errors.Wrap(err, "while creating Gardener cluster config"))
	cli, err := gardener.NewClient(clusterCfg)
	fatalOnError(errors.Wrap(err, "while creating Gardener client"))
	gardenerNamespace := fmt.Sprintf("garden-%s", cfg.Gardener.Project)
	shootClient := cli.Shoots(gardenerNamespace)

	ctx := context.Background()
	brokerClient := broker.NewClient(ctx, cfg.Broker)

	// create storage

	db, conn, err := storage.NewFromConfig(cfg.Database, log.WithField("service", "storage"))
	fatalOnError(err)
	dbStatsCollector := sqlstats.NewStatsCollector("broker", conn)
	prometheus.MustRegister(dbStatsCollector)

	svc := environmentscleanup.NewService(shootClient, brokerClient, db.Instances(), cfg.MaxAgeHours, cfg.LabelSelector)
	err = svc.PerformCleanup()
	if err != nil {
		fatalOnError(err)
	}
	log.Info("Kyma Environments cleanup performed successfully")
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
