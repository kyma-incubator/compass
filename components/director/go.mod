module github.com/kyma-incubator/compass/components/director

go 1.15

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/dlmiddlecote/sqlstats v1.0.1
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/kyma-incubator/compass/components/connector v0.0.0-20210224145945-7c0650085504
	github.com/kyma-incubator/compass/components/operations-controller v0.0.0-20210213091620-beb5492e9d8b
	github.com/lestrrat-go/jwx v0.9.2
	github.com/lib/pq v1.9.0
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/onrik/logrus v0.8.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.8
	github.com/tidwall/sjson v1.1.4
	github.com/vektah/gqlparser v1.3.1
	github.com/vrischmann/envconfig v1.3.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3
	golang.org/x/sys v0.0.0-20210124154548-22da62e12c0c // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.5.0
)
