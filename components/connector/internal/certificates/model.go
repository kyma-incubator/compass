package certificates

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/connector/config"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
)

type CSRSubject struct {
	CommonName string
	CSRSubjectConsts
}

type CSRSubjectConsts struct {
	Country            string
	Organization       string
	OrganizationalUnit string
	Locality           string
	Province           string
}

func NewCSRSubjectConsts(config config.Config) CSRSubjectConsts {
	return CSRSubjectConsts{
		Country:            config.CSRSubject.Country,
		Organization:       config.CSRSubject.Organization,
		OrganizationalUnit: config.CSRSubject.OrganizationalUnit,
		Locality:           config.CSRSubject.Locality,
		Province:           config.CSRSubject.Province,
	}
}

func (s CSRSubjectConsts) ToString(commonName string) string {
	return fmt.Sprintf("O=%s,OU=%s,L=%s,ST=%s,C=%s,CN=%s", s.Organization, s.OrganizationalUnit, s.Locality, s.Province, s.Country, commonName)
}

type EncodedCertificateChain struct {
	CertificateChain  string
	ClientCertificate string
	CaCertificate     string
}

func ToCertificationResult(encodedChain EncodedCertificateChain) externalschema.CertificationResult {
	return externalschema.CertificationResult{
		CertificateChain:  encodedChain.CertificateChain,
		ClientCertificate: encodedChain.ClientCertificate,
		CaCertificate:     encodedChain.CaCertificate,
	}
}
