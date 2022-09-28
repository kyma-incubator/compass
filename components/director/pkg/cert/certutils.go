package cert

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"regexp"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/util"

	"github.com/pkg/errors"

	"github.com/google/uuid"
)

const (
	// ConsumerTypeExtraField is the consumer type json field in the auth session body extra.
	ConsumerTypeExtraField = "consumer_type"
	// AccessLevelsExtraField is the tenant access levels json field in the auth session body extra.
	AccessLevelsExtraField = "tenant_access_levels"
	// InternalConsumerIDField is the internal consumer id json field in the auth session body extra.
	InternalConsumerIDField = "internal_consumer_id"
)

// GetOrganization returns the O part of the subject
func GetOrganization(subject string) string {
	return getRegexMatch("O=([^(,|+)]+)", subject)
}

// GetOrganizationalUnit returns the first OU of the subject
func GetOrganizationalUnit(subject string) string {
	return getRegexMatch("OU=([^(,|+)]+)", subject)
}

// GetUUIDOrganizationalUnit returns the OU that is a valid UUID or empty string if there is no OU that is a valid UUID
func GetUUIDOrganizationalUnit(subject string) string {
	orgUnits := GetAllOrganizationalUnits(subject)
	for _, orgUnit := range orgUnits {
		if _, err := uuid.Parse(orgUnit); err == nil {
			return orgUnit
		}
	}
	return ""
}

// GetRemainingOrganizationalUnit returns the OU that is remaining after matching previously expected ones based on a given pattern
func GetRemainingOrganizationalUnit(organizationalUnitPattern string, ouRegionPattern string) func(string) string {
	return func(subject string) string {
		regex := ConstructOURegex(organizationalUnitPattern, ouRegionPattern)
		orgUnitRegex := regexp.MustCompile(regex)
		orgUnits := GetAllOrganizationalUnits(subject)

		remainingOrgUnit := ""
		matchedOrgUnits := 0
		for _, orgUnit := range orgUnits {
			if !orgUnitRegex.MatchString(orgUnit) {
				remainingOrgUnit = orgUnit
			} else {
				matchedOrgUnits++
			}
		}

		return remainingOrgUnit
	}
}

func ConstructOURegex(patterns ...string) string {
	nonEmptyStr := make([]string, 0)
	for _, pattern := range patterns {
		if len(pattern) > 0 {
			nonEmptyStr = append(nonEmptyStr, pattern)
		}
	}
	return strings.Join(nonEmptyStr, "|")
}

// GetAllOrganizationalUnits returns all OU parts of the subject
func GetAllOrganizationalUnits(subject string) []string {
	return getAllRegexMatches("OU=([^(,|+)]+)", subject)
}

// GetCountry returns the C part of the subject
func GetCountry(subject string) string {
	return getRegexMatch("C=([^(,|+)]+)", subject)
}

// GetProvince returns the ST part of the subject
func GetProvince(subject string) string {
	return getRegexMatch("ST=([^(,|+)]+)", subject)
}

// GetLocality returns the L part of the subject
func GetLocality(subject string) string {
	return getRegexMatch("L=([^(,|+)]+)", subject)
}

// GetCommonName returns the CN part of the subject
func GetCommonName(subject string) string {
	return getRegexMatch("CN=([^,]+)", subject)
}

// GetAuthSessionExtra returns an appropriate auth session extra for the given consumerType, accessLevel and internalConsumerID
func GetAuthSessionExtra(consumerType, internalConsumerID string, accessLevels []string) map[string]interface{} {
	return map[string]interface{}{
		ConsumerTypeExtraField:  consumerType,
		AccessLevelsExtraField:  accessLevels,
		InternalConsumerIDField: internalConsumerID,
	}
}

// GetPossibleRegexTopLevelMatches returns the number of possible top level matches of a regex pattern.
// This means that the pattern will be inspected and split only based on the most top level '|' delimiter
// and inner group '|' delimiters won't be taken into account.
func GetPossibleRegexTopLevelMatches(pattern string) int {
	if pattern == "" {
		return 0
	}
	count := 1
	openedGroups := 0
	for _, char := range pattern {
		switch char {
		case '|':
			if openedGroups == 0 {
				count++
			}
		case '(':
			openedGroups++
		case ')':
			openedGroups--
		default:
			continue
		}
	}
	return count
}

// DecodeCertificates accepts raw certificate chain and return slice of parsed certificates
func DecodeCertificates(pemCertChain []byte) ([]*x509.Certificate, error) {
	if pemCertChain == nil {
		return nil, errors.New("Certificate data is empty")
	}

	var certificates []*x509.Certificate

	for block, rest := pem.Decode(pemCertChain); block != nil && rest != nil; {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to decode one of the pem blocks")
		}

		certificates = append(certificates, cert)

		block, rest = pem.Decode(rest)
	}

	if len(certificates) == 0 {
		return nil, errors.New("No certificates found in the pem block")
	}

	return certificates, nil
}

// NewTLSCertificate creates tls certificate from given certificate chain in form of slice of certificates
func NewTLSCertificate(key *rsa.PrivateKey, certificates ...*x509.Certificate) tls.Certificate {
	rawCerts := make([][]byte, len(certificates))
	for i, c := range certificates {
		rawCerts[i] = c.Raw
	}

	return tls.Certificate{
		Certificate: rawCerts,
		PrivateKey:  key,
	}
}

func getRegexMatch(regex, text string) string {
	matches := getAllRegexMatches(regex, text)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func getAllRegexMatches(regex, text string) []string {
	cnRegex := regexp.MustCompile(regex)
	matches := cnRegex.FindAllStringSubmatch(text, -1)

	result := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		result = append(result, match[1])
	}

	return result
}

// ParseCertificate creates a tls.Certificate from certificate and key
// The cert/key can be in PEM format or can be base64 encoded
func ParseCertificate(cert string, key string) (*tls.Certificate, error) {
	if cert == "" || key == "" {
		return nil, errors.New("The cert/key is required")
	}

	certChainBytes := util.TryDecodeBase64(cert)
	privateKeyBytes := util.TryDecodeBase64(key)

	return ParseCertificateBytes(certChainBytes, privateKeyBytes)
}

// ParseCertificateBytes creates a tls.Certificate from certificate and key
func ParseCertificateBytes(cert []byte, key []byte) (*tls.Certificate, error) {
	certs, err := DecodeCertificates(cert)
	if err != nil {
		return nil, errors.Wrap(err, "Error while decoding certificate pem block")
	}

	privateKeyPem, _ := pem.Decode(key)
	if privateKeyPem == nil {
		return nil, errors.New("Error while decoding private key pem block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPem.Bytes)
	if err != nil {
		pkcs8PrivateKey, err := x509.ParsePKCS8PrivateKey(privateKeyPem.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		privateKey, ok = pkcs8PrivateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, err
		}
	}

	tlsCert := NewTLSCertificate(privateKey, certs...)

	return &tlsCert, nil
}
