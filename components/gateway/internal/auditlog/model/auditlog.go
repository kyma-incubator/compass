package model

import "net"

type ConfigurationChange struct {
	User       string      `json:"user"`
	Object     Object      `json:"object"`
	Attributes []Attribute `json:"attributes"`
	Success    *bool       `json:"success"`
	AuditlogMetadata
}

type Attribute struct {
	Name string `json:"name"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

type SecurityEvent struct {
	User string  `json:"user"`
	IP   *net.IP `json:"ip"`
	Data string  `json:"data"`
	AuditlogMetadata
}

type AuditlogMetadata struct {
	Time   string `json:"time"`
	Tenant string `json:"tenant"`
	UUID   string `json:"uuid"`
}

type SecurityEventData struct {
	ID     map[string]string `json:"id"`
	Reason string            `json:"reason"`
}

type Object struct {
	ID   map[string]string `json:"id"`
	Type string            `json:"type"`
}

type ID struct {
	ExtraData map[string]string `json:"extra_data"`
}
