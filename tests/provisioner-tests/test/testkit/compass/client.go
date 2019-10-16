package compass

//directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
//provisionerSchema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
//"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass/director"
//"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/graphql"
//"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/oauth"
//"github.com/patrickmn/go-cache"

const (
	TenantHeader        = "Tenant"
	AuthorizationHeader = "Authorization"
)

//type ProvisionerInterface interface {
//	ProvisionRuntime(runtimeID string, config provisionerSchema.ProvisionRuntimeInput) (string, error)
//	UpgradeRuntime(runtimeID string, config provisionerSchema.UpgradeRuntimeInput) (string, error)
//	DeprovisionRuntime(runtimeID string) (string, error)
//	ReconnectRuntimeAgent(runtimeID string) (string, error)
//	RuntimeStatus(runtimeID string) (provisionerSchema.RuntimeStatus, error)
//	RuntimeOperationStatus(operationID string) (provisionerSchema.OperationStatus, error)
//}
//
//type DirectorInterface interface {
//	RegisterRuntime(input directorSchema.RuntimeInput) (director.Runtime, error)
//}
//
//type Client struct {
//	provisionerGQLClient *graphql.Client
//	directorGQLClient    *graphql.Client
//	oauthClient          *oauth.Client
//	credentials          oauth.Credentials
//	tokenCache           *cache.Cache
//}
