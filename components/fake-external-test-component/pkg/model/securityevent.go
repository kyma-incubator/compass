package model

import (
	"net"
)

type SecuritEvent struct {
	UUID   string `json:"uuid"`
	User   string `json:"user"`
	Time   string `json:"time"`
	IP     net.IP `json:"ip"`
	Data   string `json:"data"`
	Tenant string `json:"tenant"`
}
