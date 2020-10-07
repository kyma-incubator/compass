package log_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestStructuredFormatter(t *testing.T) {
	suite.Run(t, new(FormatterSuite))
}

type FormatterSuite struct {
	suite.Suite

	buffer *bytes.Buffer
	entry  *logrus.Entry
}

func (suite *FormatterSuite) SetupTest() {
	suite.buffer = &bytes.Buffer{}
	config := log.DefaultConfig()
	config.Format = "structured"
	ctx, err := log.Configure(context.TODO(), config)
	suite.Require().NoError(err)
	suite.entry = log.LoggerFromContext(ctx)
	suite.entry.Logger.SetOutput(suite.buffer)
}

func (suite *FormatterSuite) TestFormatterIncludeDefaultFields() {
	suite.entry.Info("test")

	suite.Require().Contains(suite.buffer.String(), `"correlation_id":`)
	suite.Require().Contains(suite.buffer.String(), `"component_type":`)
	suite.Require().Contains(suite.buffer.String(), `"level":"info"`)
	suite.Require().Contains(suite.buffer.String(), `"logger":`)
	suite.Require().Contains(suite.buffer.String(), `"type":"log"`)
	suite.Require().Contains(suite.buffer.String(), `"written_at":`)
	suite.Require().Contains(suite.buffer.String(), `"written_ts":`)
	suite.Require().Contains(suite.buffer.String(), `"msg":"test"`)
}

func (suite *FormatterSuite) TestFormatterAppendErrorsToTheMessage() {
	err := fmt.Errorf("error message")
	suite.entry.WithError(err).Error("test message")

	suite.Require().Contains(suite.buffer.String(), `"level":"error"`)
	suite.Require().Contains(suite.buffer.String(), `"msg":"test message: `+err.Error()+`"`)
}

func (suite *FormatterSuite) TestFormatterCustomFields() {
	suite.entry.WithField("test_field", "test_value").Info("test")

	suite.Require().Contains(suite.buffer.String(), `,"test_field":"test_value",`)
}
