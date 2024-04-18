package certsubjmapping

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/pkg/errors"
)

type SubjectConsumerTypeMapping struct {
	Subject            string   `json:"subject"`
	ConsumerType       string   `json:"consumer_type"`
	InternalConsumerID string   `json:"internal_consumer_id"`
	TenantAccessLevels []string `json:"tenant_access_levels"`
}

func (s *SubjectConsumerTypeMapping) Validate() error {
	if len(s.Subject) < 1 {
		return errors.New("subject is not provided")
	}

	if !inputvalidation.SupportedConsumerTypes[consumer.Type(s.ConsumerType)] {
		return fmt.Errorf("consumer type %s is not valid", s.ConsumerType)
	}

	for _, al := range s.TenantAccessLevels {
		if !inputvalidation.SupportedAccessLevels[al] {
			return fmt.Errorf("tenant access level %s is not valid", al)
		}
	}

	return nil
}

func UnmarshalMappings(certSubjectMappingsFromEnv string) ([]SubjectConsumerTypeMapping, error) {
	var mappingsFromEnv []SubjectConsumerTypeMapping
	if err := json.Unmarshal([]byte(certSubjectMappingsFromEnv), &mappingsFromEnv); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling mappings")
	}

	return mappingsFromEnv, nil
}
