package certificates

import (
	"fmt"

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

func (s CSRSubject) ToString() string {
	return fmt.Sprintf("O=%s,OU=%s,L=%s,ST=%s,C=%s,CN=%s", s.Organization, s.OrganizationalUnit, s.Locality, s.Province, s.Country, s.CommonName)
}

type EncodedCertificateChain struct {
	CertificateChain  string
	ClientCertificate string
	CaCertificate     string
}

func ToCertificationResult(encodedChain EncodedCertificateChain) externalschema.CertificationResult {
	return externalschema.CertificationResult{
		Certificate:       encodedChain.CertificateChain,
		ClientCertificate: encodedChain.ClientCertificate,
		CaCertificate:     encodedChain.CaCertificate,
	}
}
