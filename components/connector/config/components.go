package config

import (
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

func InitInternalComponents(cfg Config, k8sClientset kubernetes.Interface) Components {
	tokenCache := tokens.NewTokenCache(cfg.Token.ApplicationExpiration, cfg.Token.RuntimeExpiration, cfg.Token.CSRExpiration)
	tokenService := tokens.NewTokenService(tokenCache, tokens.NewTokenGenerator(cfg.Token.Length))
	revokedCertsRepository := newRevokedCertsRepository(k8sClientset, namespacedname.Parse(cfg.RevocationConfigMapName))

	authenticator := authentication.NewAuthenticator()

	secretsRepository := newSecretsRepository(k8sClientset)
	certificateUtility := certificates.NewCertificateUtility(cfg.CertificateValidityTime)
	certificateService := certificates.NewCertificateService(
		secretsRepository,
		certificateUtility,
		namespacedname.Parse(cfg.CASecretName),
		namespacedname.Parse(cfg.RootCACertificateSecretName),
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
	}
}

func newSecretsRepository(coreClientSet kubernetes.Interface) secrets.Repository {
	core := coreClientSet.CoreV1()

	return secrets.NewRepository(func(namespace string) secrets.Manager {
		return core.Secrets(namespace)
	})
}

func newRevokedCertsRepository(coreClientSet kubernetes.Interface, revocationSecret types.NamespacedName) revocation.RevocationListRepository {
	cmi := coreClientSet.CoreV1().ConfigMaps(revocationSecret.Namespace)

	return revocation.NewRepository(cmi, revocationSecret.Name)
}
