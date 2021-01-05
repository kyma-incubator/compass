module github.com/kyma-incubator/compass/components/director

go 1.13

require (
	github.com/99designs/gqlgen v0.9.3
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/agnivade/levenshtein v1.0.3 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/avast/retry-go v2.4.3+incompatible
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/dlmiddlecote/sqlstats v1.0.0
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-ozzo/ozzo-validation v3.6.0+incompatible
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/lestrrat-go/jwx v0.9.0
	github.com/lib/pq v1.2.0
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/matryer/is v1.3.0 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/onrik/logrus v0.8.0
	github.com/pkg/errors v0.9.1
	github.com/pquerna/cachecontrol v0.0.0-20201205024021-ac21108117ac // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.6.4
	github.com/vektah/gqlparser v1.3.1
	github.com/vrischmann/envconfig v1.2.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.1.1-0.20190912152152-6a016cf16650
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/tools v0.0.0-20200323144430-8dcfad9e016e // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
)

replace (
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8
)
