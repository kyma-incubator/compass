package types

type TenantInfo struct {
	CertSubject string `json:"certSubject"`
}

func (ti TenantInfo) Tenant() string {
	return ""
}
