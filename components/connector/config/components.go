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

	RevokedCertsRepository revocation.RevocationListRepository

	CertificateService certificates.Service

	SubjectConsts certificates.CSRSubjectConsts
}

func InitInternalComponents(cfg Config, k8sClientSet kubernetes.Interface) (Components, certificates.Loader, revocation.Loader) {
	tokenCache := tokens.NewTokenCache(cfg.Token.ApplicationExpiration, cfg.Token.RuntimeExpiration, cfg.Token.CSRExpiration)
	tokenService := tokens.NewTokenService(tokenCache, tokens.NewTokenGenerator(cfg.Token.Length))

	revocationListCache := revocation.NewCache()
	revocationListConfigMap := namespacedname.Parse(cfg.RevocationConfigMapName)
	revokedCertsRepository := newRevokedCertsRepository(k8sClientSet, revocationListConfigMap, revocationListCache)

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
	revocationListLoader := revocation.NewRevocationListLoader(revocationListCache,
		k8sClientSet.CoreV1().ConfigMaps(revocationListConfigMap.Namespace),
		revocationListConfigMap.Name,
		time.Second,
	)

	csrSubjectConsts := certificates.CSRSubjectConsts{
		Country:            cfg.CSRSubject.Country,
		Organization:       cfg.CSRSubject.Organization,
		OrganizationalUnit: cfg.CSRSubject.OrganizationalUnit,
		Locality:           cfg.CSRSubject.Locality,
		Province:           cfg.CSRSubject.Province,
	}

	return Components{
		TokenService:           tokenService,
		Authenticator:          authentication.NewAuthenticator(),
		RevokedCertsRepository: revokedCertsRepository,
		CertificateService:     certificateService,
		SubjectConsts:          csrSubjectConsts,
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
