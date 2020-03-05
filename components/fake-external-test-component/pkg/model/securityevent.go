package model

import (
	"net"
	"time"
)

type SecuritEvent struct {
	UUID   string    `json:"uuid"`
	User   string    `json:"user"`
	Time   time.Time `json:"time"`
	IP     net.IP    `json:"ip"`
	Data   string    `json:"data"`
	Tenant string    `json:"tenant"`
}
