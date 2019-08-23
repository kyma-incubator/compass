package authentication

import (
	"net/http"
	"regexp"
)

const ClientCertHeader = "X-Forwarded-Client-Cert"

//go:generate mockery -name=CertificateHeaderParser
type CertificateHeaderParser interface {
	GetCertificateData(r *http.Request) (string, string, bool)
}

type CertificateData struct {
	Hash       string
	CommonName string
}

type certificateInfo struct {
	Hash    string
	Subject string
}

type headerParser struct {
	country            string
	locality           string
	province           string
	organization       string
	organizationalUnit string
}

func NewHeaderParser(country, locality, province, organization, organizationalUnit string) *headerParser {
	return &headerParser{
		country:            country,
		locality:           locality,
		organization:       organization,
		organizationalUnit: organizationalUnit,
		province:           province,
	}
}

func (hp *headerParser) GetCertificateData(r *http.Request) (string, string, bool) {
	certHeader := r.Header.Get(ClientCertHeader)
	if certHeader == "" {
		return "", "", false
	}

	subjectRegex := regexp.MustCompile(`Subject="(.*?)"`)
	hashRegex := regexp.MustCompile(`Hash=([0-9a-f]*)`)

	subjects := extractFromHeader(certHeader, subjectRegex)
	hashes := extractFromHeader(certHeader, hashRegex)

	certificateInfos := createCertInfos(subjects, hashes)

	certificateInfo, found := hp.getCertificateInfoWithMatchingSubject(certificateInfos)
	if !found {
		return "", "", false
	}

	return GetCommonName(certificateInfo.Subject), certificateInfo.Hash, true
}

func createCertInfos(subjects, hashes []string) []certificateInfo {
	certInfos := make([]certificateInfo, len(subjects))
	for i := 0; i < len(subjects); i++ {
		certInfo := newCertificateInfo(subjects[i], hashes[i])
		certInfos[i] = certInfo
	}
	return certInfos
}

func (hp *headerParser) getCertificateInfoWithMatchingSubject(infos []certificateInfo) (certificateInfo, bool) {
	for _, info := range infos {
		if hp.isSubjectMatching(info) {
			return info, true
		}
	}
	return certificateInfo{}, false
}

func newCertificateInfo(subject, hash string) certificateInfo {
	certInfo := certificateInfo{
		Hash:    hash,
		Subject: subject,
	}
	return certInfo
}

func (hp *headerParser) isSubjectMatching(i certificateInfo) bool {
	return GetOrganization(i.Subject) == hp.organization && GetOrganizationalUnit(i.Subject) == hp.organizationalUnit &&
		GetCountry(i.Subject) == hp.country && GetLocality(i.Subject) == hp.locality && GetProvince(i.Subject) == hp.province
}

func extractFromHeader(certHeader string, regex *regexp.Regexp) []string {
	var matchedStrings []string

	matches := regex.FindAllStringSubmatch(certHeader, -1)

	for _, match := range matches {
		hash := get(match, 1)
		matchedStrings = append(matchedStrings, hash)
	}

	return matchedStrings
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}
