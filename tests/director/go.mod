module github.com/kyma-incubator/compass/tests/director

go 1.15

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible
	github.com/google/uuid v1.2.0
	github.com/kyma-incubator/compass/components/director v0.0.0-20210304180146-8ebee1d105ad
	github.com/kyma-incubator/compass/components/external-services-mock v0.0.0-20210305150713-2817870bb301
	github.com/kyma-incubator/compass/components/gateway v0.0.0-20210203135116-086a057e4d3c
	github.com/kyma-incubator/compass/components/operations-controller v0.0.0-20210304094417-deb791b48d6a
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/vrischmann/envconfig v1.3.0
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.2
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
)
