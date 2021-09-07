package log

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type kibanaEntry struct {
	WrittenAt        string        `json:"written_at"`
	WrittenTimestamp string        `json:"written_ts"`
	Type             string        `json:"type"`
	Logger           string        `json:"logger"`
	Level            string        `json:"level"`
	Message          string        `json:"msg"`
	Fields           logrus.Fields `json:"-"`
}

// MarshalJSON marshals the kibana entry by inlining the logrus fields instead of being nested in the "Fields" tag
func (k kibanaEntry) MarshalJSON() ([]byte, error) {
	type Entry kibanaEntry
	bytes, err := json.Marshal(Entry(k))
	if err != nil {
		return nil, err
	}

	var result map[string]json.RawMessage
	if err = json.Unmarshal(bytes, &result); err != nil {
		return nil, err
	}

	for k, v := range k.Fields {
		field, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		result[k] = field
	}

	return json.Marshal(result)
}

// KibanaFormatter is a logrus formatter that formats an entry for Kibana
type KibanaFormatter struct {
}

// Format formats a logrus entry for Kibana logging
func (f *KibanaFormatter) Format(e *logrus.Entry) ([]byte, error) {
	componentName, exists := e.Data[fieldComponentName].(string)
	if !exists {
		componentName = "-"
	}
	delete(e.Data, fieldComponentName)
	if errorField, exists := e.Data[logrus.ErrorKey].(error); exists {
		e.Message = e.Message + ": " + errorField.Error()
	}
	delete(e.Data, logrus.ErrorKey)

	kibanaEntry := &kibanaEntry{
		Logger:           componentName,
		Level:            e.Level.String(),
		Message:          e.Message,
		Type:             "log",
		WrittenAt:        e.Time.UTC().Format(time.RFC3339Nano),
		WrittenTimestamp: fmt.Sprintf("%d", e.Time.UTC().Unix()),
		Fields:           e.Data,
	}
	serialized, err := json.Marshal(kibanaEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON: %v", err)
	}
	return append(serialized, '\n'), nil
}
