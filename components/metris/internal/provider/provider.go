package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/metris/internal/gardener"
	"github.com/kyma-incubator/compass/components/metris/internal/provider/azure"
	"go.uber.org/zap"
)

type CloudProviderType int

const (
	AWS CloudProviderType = iota
	AZURE
	GCP
)

type Config struct {
	Type             string        `kong:"help='Provider to fetch metrics from. (azure)',enum='azure',env='PROVIDER_TYPE',required=true,default='azure',hidden=true"`
	PollInterval     time.Duration `kong:"help='Interval at which metrics are fetch.',env='PROVIDER_POLLINTERVAL',required=true,default='1m'"`
	Workers          int           `kong:"help='Number of workers to fetch metrics.',env='PROVIDER_WORKERS',required=true,default=10"`
	Buffer           int           `kong:"help='Number of accounts that the buffer can have.',env='PROVIDER_BUFFER',required=true,default=100"`
	ClientTraceLevel int           `kong:"help='Provider client trace level (0=disabled, 1=headers, 2=body)',env='PROVIDER_CLIENT_TRACE_LEVEL',default=0"`
}

// Provider interface contains all behaviors for a provider.
type Provider interface {
	Collect(ctx context.Context, wg *sync.WaitGroup)
}

// NewProvider Create a new provider from the name.
func NewProvider(config *Config, accountsChannel <-chan *gardener.Account, eventsChannel chan<- *[]byte, logger *zap.SugaredLogger) (Provider, error) {
	logger = logger.With("component", config.Type)

	switch strings.ToLower(config.Type) {
	case AZURE.String():
		logger.Debug("initializing provider")
		return azure.NewAzureProvider(config.Workers, config.PollInterval, accountsChannel, eventsChannel, logger, config.ClientTraceLevel), nil
	default:
		return nil, fmt.Errorf("provider %s undefined", config.Type)
	}
}

func getSupportedProviders() []string {
	return []string{"aws", "azure", "gcp"}
}

func (p CloudProviderType) String() string {
	return getSupportedProviders()[p]
}
