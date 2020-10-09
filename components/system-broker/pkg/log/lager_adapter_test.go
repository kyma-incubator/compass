package log_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"code.cloudfoundry.org/lager"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestLagerAdapter_Session(t *testing.T) {
	session := "test-session"
	adapter := log.NewDefaultLagerAdapter()
	require.NotNil(t, adapter)
	logger := adapter.Session(session)
	require.Contains(t, logger.SessionName(), session)
}

func TestLagerAdapter(t *testing.T) {
	suite.Run(t, new(LagerAdapterSuite))
}

type LagerAdapterSuite struct {
	suite.Suite

	buffer        *bytes.Buffer
	adapter       *log.LagerAdapter
	data          lager.Data
	inheritedData lager.Data
}

func (suite *LagerAdapterSuite) SetupTest() {
	suite.buffer = &bytes.Buffer{}
	config := log.DefaultConfig()
	config.Format = "structured"
	_, err := log.Configure(context.TODO(), config)
	suite.Require().NoError(err)

	log.D().Logger.SetOutput(suite.buffer)

	suite.adapter = log.NewDefaultLagerAdapter()
	suite.Require().NotNil(suite.adapter)

	suite.data = lager.Data{
		"a": "b",
	}
	suite.inheritedData = lager.Data{
		"b": "a",
	}
}

func (suite *LagerAdapterSuite) TestLagerAdapterLogsAllMandatoryProps() {
	suite.adapter.Info("test", suite.data)

	suite.Require().Contains(suite.buffer.String(), `"correlation_id":`)
	suite.Require().Contains(suite.buffer.String(), `"msg":"test"`)
	suite.Require().Contains(suite.buffer.String(), `"a":"b"`)
}

func (suite *LagerAdapterSuite) TestLagerAdapterLogsAllInheritedDataWhenNewSessionIsEstablished() {
	newAdapter := suite.adapter.Session("session", suite.inheritedData)
	newAdapter.Info("test")

	suite.Require().Contains(suite.buffer.String(), `"correlation_id":`)
	suite.Require().Contains(suite.buffer.String(), `"msg":"test"`)
	suite.Require().Contains(suite.buffer.String(), `"b":"a"`)
}

func (suite *LagerAdapterSuite) TestLagerAdapterAppendErrorsToTheMessage() {
	suite.adapter.Error("test", errors.New("test-error"), suite.data)

	suite.Require().Contains(suite.buffer.String(), `"correlation_id":`)
	suite.Require().Contains(suite.buffer.String(), `"msg":"test: test-error"`)
	suite.Require().Contains(suite.buffer.String(), `"a":"b"`)
}

func (suite *LagerAdapterSuite) TestLagerAdapterDoesNotUseSinks() {
	suite.adapter.RegisterSink(nil)

	suite.Require().Contains(suite.buffer.String(), "LagerAdapter does not work with sinks.")
}
