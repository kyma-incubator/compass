package log

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKibanaFormatter(t *testing.T) {

	buffer := &bytes.Buffer{}
	ctx, err := Configure(context.TODO(), &Config{
		Level:                  "debug",
		Format:                 "kibana",
		BootstrapCorrelationID: "bootstrap",
		Output:                 os.Stdout.Name(),
	})
	require.NoError(t, err)
	entry := LoggerFromContext(ctx)
	entry.Logger.SetOutput(buffer)
	entry.Debug("test")

	t.Run("should contain log level", func(t *testing.T) {
		assert.Contains(t, buffer.String(), `"level":"debug"`)
	})

	t.Run("should contain logger information", func(t *testing.T) {
		assert.Contains(t, buffer.String(), `"logger":`)
	})

	t.Run("should contain log type", func(t *testing.T) {
		assert.Contains(t, buffer.String(), `"type":"log"`)
	})

	t.Run("should contain written_at human readable timestamp", func(t *testing.T) {
		assert.Contains(t, buffer.String(), `"written_at":`)
	})

	t.Run("should contain written_ts timestamp", func(t *testing.T) {
		assert.Contains(t, buffer.String(), `"written_ts":`)
	})

	t.Run("should contain the message", func(t *testing.T) {
		assert.Contains(t, buffer.String(), `"msg":"test"`)
	})

	t.Run("error is logged", func(t *testing.T) {
		err := fmt.Errorf("error message")
		entry.WithError(err).Error("test message")
		assert.Contains(t, buffer.String(), `"level":"error"`)
		assert.Contains(t, buffer.String(), `"msg":"test message: `+err.Error()+`"`)
	})

	t.Run("custom field is logged", func(t *testing.T) {
		entry.WithField("test_field", "test_value").Debug("test")
		assert.Contains(t, buffer.String(), `,"test_field":"test_value",`)
	})
}
