package inputvalidation

import (
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/pkg/errors"
	"strings"
)

type certSubjectValidator struct{}

var IsValidCertSubject = &certSubjectValidator{}

func (v *certSubjectValidator) Validate(value interface{}) error {
	s, isNil, err := ensureIsString(value)
	if err != nil {
		return err
	}
	if isNil {
		return nil
	}

	expectedSubjectComponents := strings.Split(s, ",")
	if len(expectedSubjectComponents) < 5 { // 5 because that's the number of certificate relative distinguished names that we expect - CountryName(C), Organization(O), OrganizationalUnit(OU), Locality(L) and CommonName(CN)
		return errors.Errorf("the number of certificate attributes are different than the expected ones. We got: %d and we need at least 5 - C, O, OU, L and CN", len(expectedSubjectComponents))
	}

	if country := cert.GetCountry(s); country == "" {
		return errors.New("missing Country property in the subject")
	}

	if org := cert.GetOrganization(s); org == "" {
		return errors.New("missing Organization property in the subject")
	}

	OUs := cert.GetAllOrganizationalUnits(s)
	if len(OUs) < 1 {
		return errors.New("missing Organization Unit property in the subject")
	}

	if locality := cert.GetLocality(s); locality == "" {
		return errors.New("missing Locality property in the subject")
	}

	if cm := cert.GetCommonName(s); cm == "" {
		return errors.New("missing Common Name property in the subject")
	}

	return nil
}
