package config

import (
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/namespacedname"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type Components struct {
	TokenCache    tokens.Cache
	TokenService  tokens.Service
	Authenticator authentication.Authenticator

	RevocationsRepository revocation.RevocationListRepository
	SecretsRepository     secrets.Repository

	CertificateUtility certificates.CertificateUtility
	CertificateService certificates.Service

	SubjectConsts certificates.CSRSubjectConsts
}

func InitInternalComponents(cfg Config, k8sClientset kubernetes.Interface) (Components, certificates.Loader, revocation.Loader) {
	tokenCache := tokens.NewTokenCache(cfg.Token.ApplicationExpiration, cfg.Token.RuntimeExpiration, cfg.Token.CSRExpiration)
	tokenService := tokens.NewTokenService(tokenCache, tokens.NewTokenGenerator(cfg.Token.Length))
	revocationListCache := revocation.NewCache()
	configmapNamespace := namespacedname.Parse(cfg.RevocationConfigMapName)
	revokedCertsRepository := newRevokedCertsRepository(k8sClientset, configmapNamespace, revocationListCache)

	authenticator := authentication.NewAuthenticator()

	caSecretName := namespacedname.Parse(cfg.CASecret.Name)
	rootCASecretName := namespacedname.Parse(cfg.RootCASecret.Name)

	certCache := certificates.NewCertificateCache()

	secretsRepository := newSecretsRepository(k8sClientset)
	certificateUtility := certificates.NewCertificateUtility(cfg.CertificateValidityTime)
	certificateService := certificates.NewCertificateService(
		certCache,
		certificateUtility,
		caSecretName.Name,
		rootCASecretName.Name,
		cfg.CASecret.CertificateKey,
		cfg.CASecret.KeyKey,
		cfg.RootCASecret.CertificateKey,
	)

	certLoader := certificates.NewCertificateLoader(certCache, secretsRepository, caSecretName, rootCASecretName)
	revocationListLoader := revocation.NewRevocationListLoader(revocationListCache,
		k8sClientset.CoreV1().ConfigMaps(configmapNamespace.Namespace),
		configmapNamespace.Name,
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
		TokenCache:            tokenCache,
		TokenService:          tokenService,
		Authenticator:         authenticator,
		RevocationsRepository: revokedCertsRepository,
		SecretsRepository:     secretsRepository,
		CertificateUtility:    certificateUtility,
		CertificateService:    certificateService,
		SubjectConsts:         csrSubjectConsts,
	}, certLoader, revocationListLoader
}

func newSecretsRepository(coreClientSet kubernetes.Interface) secrets.Repository {
	core := coreClientSet.CoreV1()

	return secrets.NewRepository(func(namespace string) secrets.Manager {
		return core.Secrets(namespace)
	})
}

func newRevokedCertsRepository(coreClientSet kubernetes.Interface, revocationSecret types.NamespacedName, cache revocation.Cache) revocation.RevocationListRepository {
	cmi := coreClientSet.CoreV1().ConfigMaps(revocationSecret.Namespace)

	return revocation.NewRepository(cmi, cache, revocationSecret.Name)
}
