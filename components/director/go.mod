module github.com/kyma-incubator/compass/components/director

go 1.13

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dlmiddlecote/sqlstats v1.0.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-multierror v1.0.0
	github.com/jmoiron/sqlx v1.2.0
	github.com/kyma-incubator/compass/components/connectivity-adapter v0.0.0-20201020133117-754956070c89
	github.com/lestrrat-go/jwx v0.9.0
	github.com/lib/pq v1.2.0
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/gjson v1.6.0
	github.com/tidwall/pretty v1.0.1 // indirect
	github.com/vektah/gqlparser v1.3.1
	github.com/vrischmann/envconfig v1.2.0
	github.com/xeipuuv/gojsonschema v1.1.1-0.20190912152152-6a016cf16650
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20201020230747-6e5568b54d1a // indirect
	golang.org/x/tools v0.0.0-20201021000207-d49c4edd7d96 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	k8s.io/apimachinery v0.17.3
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
