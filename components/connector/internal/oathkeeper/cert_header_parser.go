package oathkeeper

import (
	"net/http"
	"regexp"

	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
)

const ClientCertHeader = "X-Forwarded-Client-Cert"

//go:generate mockery -name=CertificateHeaderParser
type CertificateHeaderParser interface {
	GetCertificateData(r *http.Request) (string, string, bool)
}

type certificateInfo struct {
	Hash    string
	Subject string
}

type headerParser struct {
	certificates.CSRSubjectConsts
}

func NewHeaderParser(csrSubjectConsts certificates.CSRSubjectConsts) CertificateHeaderParser {
	return &headerParser{
		CSRSubjectConsts: csrSubjectConsts,
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
	if len(subjects) != len(hashes) {
		return []certificateInfo{}
	}

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
	return GetOrganization(i.Subject) == hp.Organization && GetOrganizationalUnit(i.Subject) == hp.OrganizationalUnit &&
		GetCountry(i.Subject) == hp.Country && GetLocality(i.Subject) == hp.Locality && GetProvince(i.Subject) == hp.Province
}

func extractFromHeader(certHeader string, regex *regexp.Regexp) []string {
	var matchedStrings []string

	matches := regex.FindAllStringSubmatch(certHeader, -1)

	for _, match := range matches {
		value := get(match, 1)
		matchedStrings = append(matchedStrings, value)
	}

	return matchedStrings
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}
	return ""
}
