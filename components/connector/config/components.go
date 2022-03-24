package config

import (
	"github.com/kyma-incubator/compass/components/connector/internal/namespacedname"

	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"k8s.io/client-go/kubernetes"
)

type Components struct {
	TokenService  tokens.Service
	Authenticator authentication.Authenticator

	CertificateService     certificates.Service
	RevokedCertsRepository revocation.RevokedCertificatesRepository

	CSRSubjectConsts certificates.CSRSubjectConsts
}

func InitInternalComponents(cfg Config, k8sClientSet kubernetes.Interface, directorGCLI tokens.GraphQLClient) (Components, certificates.Loader) {
	caSecret := namespacedname.Parse(cfg.CASecret.Name)

	rootCASecret := namespacedname.Parse(cfg.RootCASecret.Name)

	certsCache := certificates.NewCertificateCache()
	certsService := certificates.NewCertificateService(
		certsCache,
		certificates.NewCertificateUtility(cfg.CertificateValidityTime),
		caSecret.Name,
		rootCASecret.Name,
		cfg.CASecret.CertificateKey,
		cfg.CASecret.KeyKey,
		cfg.RootCASecret.CertificateKey,
	)
	certsLoader := certificates.NewCertificateLoader(certsCache, newSecretsRepository(k8sClientSet), caSecret, rootCASecret)

	revokedCertsConfigMap := namespacedname.Parse(cfg.RevocationConfigMapName)

	revokedCertsRepository := newRevokedCertsRepository(k8sClientSet, revokedCertsConfigMap)

	return Components{
		Authenticator:          authentication.NewAuthenticator(),
		TokenService:           tokens.NewTokenService(directorGCLI),
		CertificateService:     certsService,
		RevokedCertsRepository: revokedCertsRepository,
		CSRSubjectConsts:       newCSRSubjectConsts(cfg),
	}, certsLoader
}

func newRevokedCertsRepository(k8sClientSet kubernetes.Interface, revokedCertsConfigMap types.NamespacedName) revocation.RevokedCertificatesRepository {
	cmi := k8sClientSet.CoreV1().ConfigMaps(revokedCertsConfigMap.Namespace)

	return revocation.NewRepository(cmi, revokedCertsConfigMap.Name)
}

func newSecretsRepository(k8sClientSet kubernetes.Interface) secrets.Repository {
	core := k8sClientSet.CoreV1()

	return secrets.NewRepository(func(namespace string) secrets.Manager {
		return core.Secrets(namespace)
	})
}

func newCSRSubjectConsts(config Config) certificates.CSRSubjectConsts {
	return certificates.CSRSubjectConsts{
		Country:            config.CSRSubject.Country,
		Organization:       config.CSRSubject.Organization,
		OrganizationalUnit: config.CSRSubject.OrganizationalUnit,
		Locality:           config.CSRSubject.Locality,
		Province:           config.CSRSubject.Province,
	}
}
