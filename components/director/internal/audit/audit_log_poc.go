package audit

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

var TENANTS = [...]string{"b27258d6-90ec-4187-85c3-2303a92a8f7a",
	"12957a43-abca-418d-8a2d-0b44d28ce3eb",
	"3dd35497-17b8-4a54-b617-7f26d5243a9f"}

var EVENTS = [...]string{
	"Created Application", "Deleted Application", "Created Runtime", "Deleted Runtime",
}

type AuditLogTest struct {
	executionFunc func(stopCh <-chan struct{})
}

func NewAuditLogTest() *AuditLogTest {
	return &AuditLogTest{}
}

func (alt *AuditLogTest) LogAuditEvent() {
	size := len(TENANTS)
	tenant := TENANTS[rand.Intn(size)]

	size = len(EVENTS)
	event := EVENTS[rand.Intn(size)]

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.WithField("tenant", tenant).
		WithField("audit", true).
		Info(event)

}
