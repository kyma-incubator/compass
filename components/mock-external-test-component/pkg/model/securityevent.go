package model

import (
	"net"
)

type SecurityEvent struct {
	UUID   string  `json:"uuid"`
	User   string  `json:"user"`
	Time   string  `json:"time"`
	IP     *net.IP `json:"ip"`
	Data   string  `json:"data"`
	Tenant string  `json:"tenant"`
}
