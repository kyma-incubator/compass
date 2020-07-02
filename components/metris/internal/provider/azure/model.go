package azure

import (
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2019-06-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/kyma-project/control-plane/components/metris/internal/gardener"
	"go.uber.org/zap"
	"k8s.io/client-go/util/workqueue"
)

type Provider struct {
	mu               sync.RWMutex
	workers          int
	pollinterval     time.Duration
	accountsChannel  chan *gardener.Account
	eventsChannel    chan<- *[]byte
	queue            workqueue.DelayingInterface
	clients          map[string]*Client
	logger           *zap.SugaredLogger
	clientTraceLevel int
}

// Client holds azure clients configuration
type Client struct {
	Account             *gardener.Account
	SubscriptionID      string
	Location            string
	logger              *zap.SugaredLogger
	computeBaseClient   *compute.BaseClient
	networkBaseClient   *network.BaseClient
	insightsBaseClient  *insights.BaseClient
	resourcesBaseClient *resources.BaseClient
	eventhubBaseClient  *eventhub.BaseClient
}

type ResourceSkus struct {
	mu   sync.Mutex
	skus map[string]*compute.ResourceSku
}

// SecretMap is a structure to decode and map kubernetes secret data values to azure client configuration
type SecretMap struct {
	ClientID        string `mapstructure:"clientID"`
	ClientSecret    string `mapstructure:"clientSecret"`
	TenantID        string `mapstructure:"tenantID"`
	SubscriptionID  string `mapstructure:"subscriptionID"`
	EnvironmentName string
}

type VMType struct {
	Name  string `json:"name"`
	Count uint32 `json:"count"`
}

type ProvisionedVolume struct {
	SizeGBTotal   uint32 `json:"size_gb_total"`
	SizeGBRounded uint32 `json:"size_gb_rounded"`
	Count         uint32 `json:"count"`
}

type Compute struct {
	VMTypes            []VMType          `json:"vm_types"`
	ProvisionedRAMGB   float64           `json:"provisioned_ram_gb"`
	ProvisionedVolumes ProvisionedVolume `json:"provisioned_volumes"`
	ProvisionedCpus    uint32            `json:"provisioned_cpus"`
}

type Networking struct {
	ProvisionedLoadBalancers uint32 `json:"provisioned_loadbalancers"`
	ProvisionedVnets         uint32 `json:"provisioned_vnets"`
	ProvisionedIps           uint32 `json:"provisioned_ips"`
}

type EventHub struct {
	NumberNamespaces     uint32  `json:"number_namespaces"`
	IncomingRequestsPT1M float64 `json:"incoming_requests_pt1m"`
	MaxIncomingBytesPT1M float64 `json:"max_incoming_bytes_pt1m"`
	MaxOutgoingBytesPT1M float64 `json:"max_outgoing_bytes_pt1m"`
	IncomingRequestsPT5M float64 `json:"incoming_requests_pt5m"`
	MaxIncomingBytesPT5M float64 `json:"max_incoming_bytes_pt5m"`
	MaxOutgoingBytesPT5M float64 `json:"max_outgoing_bytes_pt5m"`
}

type Event struct {
	Timestamp      string     `json:"timestamp"`
	ResourceGroups []string   `json:"resource_groups"`
	Compute        Compute    `json:"compute"`
	Networking     Networking `json:"networking"`
	EventHub       EventHub   `json:"event_hub"`
}

type Events map[string][]*Event
