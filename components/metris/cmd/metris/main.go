package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/kyma-incubator/compass/components/metris/internal/edp"
	"github.com/kyma-incubator/compass/components/metris/internal/gardener"
	"github.com/kyma-incubator/compass/components/metris/internal/provider"
	"github.com/kyma-incubator/compass/components/metris/internal/server"
	"github.com/kyma-incubator/compass/components/metris/internal/utils"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// version is the current version, set by the go linker's -X flag at build time.
	version = "dev"
)

type app struct {
	Version        kong.VersionFlag `kong:"help='Print version information and quit.'"`
	ConfigFile     kong.ConfigFlag  `kong:"help='Location of the config file.',type='path'"`
	LogLevel       string           `kong:"help='Logging level. (${enum})',enum='${loglevels}',default='info',env='METRIS_LOGLEVEL'"`
	Kubeconfig     string           `kong:"help='Path to the Gardener kubeconfig file.',required=true,default='${kubeconfig}',env='METRIS_KUBECONFIG'"`
	ServerConfig   server.Config    `kong:"embed=true"`
	ProviderConfig provider.Config  `kong:"embed=true,prefix='provider-'"`
	EDPConfig      edp.Config       `kong:"embed=true,prefix='edp-'"`
}

func main() {
	var (
		err        error
		homefld    string
		kubeconfig string
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	homefld, err = homedir.Dir()
	if err == nil {
		kubeconfig = filepath.Join(homefld, ".kube", "config")
	}

	cli := app{}
	clictx := kong.Parse(&cli,
		kong.Name("metris"),
		kong.Description("Metris is a metering component that collects data and sends them to EDP."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: false}),
		kong.Vars{
			"version":    fmt.Sprintf("Version: %s\n", version),
			"kubeconfig": kubeconfig,
			"loglevels":  "debug,info,warn,error",
		},
		kong.Configuration(kong.JSON, ""),
	)

	var logger *zap.SugaredLogger
	{
		var loglevel zapcore.Level
		err = loglevel.UnmarshalText([]byte(cli.LogLevel))
		if err != nil {
			loglevel = zapcore.InfoLevel
		}

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder

		encoder := zapcore.NewConsoleEncoder(encoderConfig)
		core := zapcore.NewCore(encoder, zapcore.Lock(os.Stderr), loglevel)
		logger = zap.New(core).Sugar()
	}

	eventsChannel := make(chan *[]byte, cli.EDPConfig.Buffer)
	accountsChannel := make(chan *gardener.Account, cli.ProviderConfig.Buffer)

	wg := sync.WaitGroup{}

	edpclient := edp.NewClient(&cli.EDPConfig, nil, eventsChannel, logger)

	go edpclient.Run(ctx, &wg)

	// start provider to begin fetching metrics from account sent by the controller
	pro, err := provider.NewProvider(&cli.ProviderConfig, accountsChannel, eventsChannel, logger)
	clictx.FatalIfErrorf(err)

	go pro.Collect(ctx, &wg)

	// start gardener controller to watch on shoots and secrets changes and send them to the provider to process
	gclient, err := gardener.NewClient(cli.Kubeconfig)
	clictx.FatalIfErrorf(err)

	ctrl, err := gardener.NewController(gclient, cli.ProviderConfig.Type, accountsChannel, logger)
	clictx.FatalIfErrorf(err)

	go ctrl.Run(ctx, &wg)

	// start web server for metris metrics and profiling
	if len(cli.ServerConfig.ListenAddr) > 0 {
		s, err := server.NewServer(cli.ServerConfig, logger)
		clictx.FatalIfErrorf(err)

		go s.Start(ctx, &wg)
	}

	utils.SetupSignalHandler(func() {
		cancel()
	})

	wg.Wait()

	logger.Info("metris stopped")

	if loggererr := logger.Sync(); loggererr != nil {
		fmt.Println(loggererr)
	}
}
