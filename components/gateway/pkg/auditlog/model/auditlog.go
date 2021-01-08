package model

import "net"

const (
	LogFormatDate = "2006-01-02T15:04:05.999Z"
)

type ConfigurationChange struct {
	User       string      `json:"user"`
	Object     Object      `json:"object"`
	Attributes []Attribute `json:"attributes"`
	Success    *bool       `json:"success"`
	Metadata
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
	Metadata
}

type Metadata struct {
	Time   string `json:"time"`
	Tenant string `json:"tenant"`
	UUID   string `json:"uuid"`
}

type SecurityEventData struct {
	ID            map[string]string `json:"id"`
	CorrelationID string            `json:"correlation_id"`
	Reason        []ErrorMessage    `json:"reason"`
}

type Object struct {
	ID   map[string]string `json:"id"`
	Type string            `json:"type"`
}

type ID struct {
	ExtraData map[string]string `json:"extra_data"`
}
