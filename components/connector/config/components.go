package config

import (
	"k8s.io/apimachinery/pkg/types"
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/namespacedname"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"k8s.io/client-go/kubernetes"
)

type Components struct {
	TokenService  tokens.Service
	Authenticator authentication.Authenticator

	CertificateService     certificates.Service
	RevokedCertsRepository revocation.RevocationListRepository

	CSRSubjectConsts certificates.CSRSubjectConsts
}

func InitInternalComponents(cfg Config, k8sClientSet kubernetes.Interface) (Components, certificates.Loader, revocation.Loader) {
	caSecret := namespacedname.Parse(cfg.CASecret.Name)
	rootCASecret := namespacedname.Parse(cfg.RootCASecret.Name)

	certCache := certificates.NewCertificateCache()
	certificateService := certificates.NewCertificateService(
		certCache,
		certificates.NewCertificateUtility(cfg.CertificateValidityTime),
		caSecret.Name,
		rootCASecret.Name,
		cfg.CASecret.CertificateKey,
		cfg.CASecret.KeyKey,
		cfg.RootCASecret.CertificateKey,
	)
	certLoader := certificates.NewCertificateLoader(certCache, newSecretsRepository(k8sClientSet), caSecret, rootCASecret)

	revocationListCache := revocation.NewCache()
	revocationListConfigMap := namespacedname.Parse(cfg.RevocationConfigMapName)
	revokedCertsRepository := newRevokedCertsRepository(k8sClientSet, revocationListConfigMap, revocationListCache)

	revocationListLoader := revocation.NewRevocationListLoader(revocationListCache,
		k8sClientSet.CoreV1().ConfigMaps(revocationListConfigMap.Namespace),
		revocationListConfigMap.Name,
		time.Second,
	)

	return Components{
		TokenService: tokens.NewTokenService(
			tokens.NewTokenCache(cfg.Token.ApplicationExpiration, cfg.Token.RuntimeExpiration, cfg.Token.CSRExpiration),
			tokens.NewTokenGenerator(cfg.Token.Length)),
		Authenticator:          authentication.NewAuthenticator(),
		RevokedCertsRepository: revokedCertsRepository,
		CertificateService:     certificateService,
		CSRSubjectConsts:       newCSRSubjectConsts(cfg),
	}, certLoader, revocationListLoader
}

func newRevokedCertsRepository(k8sClientSet kubernetes.Interface, revocationListConfigMap types.NamespacedName, cache revocation.Cache) revocation.RevocationListRepository {
	cmi := k8sClientSet.CoreV1().ConfigMaps(revocationListConfigMap.Namespace)

	return revocation.NewRepository(cmi, revocationListConfigMap.Name, cache)
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
