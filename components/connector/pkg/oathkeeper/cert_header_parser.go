package oathkeeper

import (
	"net/http"
	"regexp"
)

//go:generate mockery --name=CertificateHeaderParser
type CertificateHeaderParser interface {
	GetCertificateData(r *http.Request) (string, string, bool)
	GetIssuer() string
}

type certificateInfo struct {
	Hash    string
	Subject string
}

type headerParser struct {
	certHeaderName         string
	issuer                 string
	isSubjectMatching      func(subject string) bool
	getClientIDFromSubject func(subject string) string
}

func NewHeaderParser(certHeaderName, issuer string, isSubjectMatching func(subject string) bool, getClientIDFromSubject func(subject string) string) *headerParser {
	return &headerParser{
		certHeaderName:         certHeaderName,
		issuer:                 issuer,
		isSubjectMatching:      isSubjectMatching,
		getClientIDFromSubject: getClientIDFromSubject,
	}
}

func (hp *headerParser) GetCertificateData(r *http.Request) (string, string, bool) {
	certHeader := r.Header.Get(hp.certHeaderName)
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

	return hp.getClientIDFromSubject(certificateInfo.Subject), certificateInfo.Hash, true
}

func (hp *headerParser) GetIssuer() string {
	return hp.issuer
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
		if hp.isSubjectMatching(info.Subject) {
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
