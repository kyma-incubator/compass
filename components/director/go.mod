module github.com/kyma-incubator/compass/components/director

go 1.17

require (
	github.com/99designs/gqlgen v0.11.0
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/dlmiddlecote/sqlstats v1.0.2
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-openapi/runtime v0.19.31
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/google/go-cmp v0.5.6
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jmoiron/sqlx v1.3.4
	github.com/kyma-incubator/compass/components/connector v0.0.0-20210922140315-d2f37c2e160d
	github.com/kyma-incubator/compass/components/operations-controller v0.0.0-20210922140924-1dc663c3ed24
	github.com/kyma-incubator/compass/components/system-broker v0.0.0-20210922132333-3deb7cf90637
	github.com/lestrrat-go/iter v1.0.1
	github.com/lestrrat-go/jwx v1.2.6
	github.com/lib/pq v1.10.3
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/maxbrunsfeld/counterfeiter/v6 v6.4.1
	github.com/mitchellh/hashstructure/v2 v2.0.2
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/onrik/logrus v0.9.0
	github.com/ory/hydra-client-go v1.10.6
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.9.1
	github.com/tidwall/sjson v1.2.2
	github.com/vektah/gqlparser/v2 v2.1.0
	github.com/vrischmann/envconfig v1.3.0
	github.com/xeipuuv/gojsonschema v1.2.0
	go.mozilla.org/pkcs7 v0.0.0-20210826202110-33d05740a352
	golang.org/x/mod v0.5.0
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	k8s.io/api v0.20.2 // DO NOT BUMP
	k8s.io/apimachinery v0.20.2 // DO NOT BUMP
	k8s.io/client-go v0.20.2 // DO NOT BUMP
	sigs.k8s.io/controller-runtime v0.8.3 // DO NOT BUMP
)
