package tenantfetcher_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testProvider           = "test"
	testFieldName          = "name"
	testFieldID            = "id"
	testFieldDiscriminator = "discriminator"
	testValueDiscriminator = "default"
)

var (
	testMapping = tenantfetcher.TenantFieldMapping{
		NameField:          testFieldName,
		IDField:            testFieldID,
		DiscriminatorField: "",
		DiscriminatorValue: "",
	}
)

func TestConverter_EventsToTenants(t *testing.T) {
	t.Run("Success when converting empty slice", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.CreatedEventsType
		in := []tenantfetcher.Event{}

		// WHEN
		out := conv.EventsToTenants(eventsType, in)

		// THEN
		assert.Empty(t, out)
	})

	t.Run("Success when converting slice with nils", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.CreatedEventsType
		in := []tenantfetcher.Event{nil, nil, nil}

		// WHEN
		out := conv.EventsToTenants(eventsType, in)

		// THEN
		assert.Empty(t, out)
	})

	t.Run("Success when converting created events (without discriminator)", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.CreatedEventsType
		in := []tenantfetcher.Event{
			fixEvent("1", "foo", testMapping),
			fixEvent("2", "bar", testMapping),
			fixEvent("3", "baz", testMapping),
		}
		expected := []model.BusinessTenantMappingInput{
			{
				Name:           "foo",
				ExternalTenant: "1",
				Provider:       testProvider,
			},
			{
				Name:           "bar",
				ExternalTenant: "2",
				Provider:       testProvider,
			},
			{
				Name:           "baz",
				ExternalTenant: "3",
				Provider:       testProvider,
			},
		}

		// WHEN
		out := conv.EventsToTenants(eventsType, in)

		// THEN
		assert.Len(t, out, len(in))
		assert.ElementsMatch(t, expected, out)
	})

	t.Run("Success when converting created events (with discriminator)", func(t *testing.T) {
		// GIVEN
		mapping := testMapping
		mapping.DiscriminatorField = testFieldDiscriminator
		mapping.DiscriminatorValue = testValueDiscriminator

		conv := tenantfetcher.NewConverter(testProvider, mapping)

		eventsType := tenantfetcher.CreatedEventsType
		in := []tenantfetcher.Event{
			fixEventWithDiscriminator("1", "foo", mapping.DiscriminatorValue, mapping),
			fixEventWithDiscriminator("2", "error", "ignore_me", mapping),
			fixEventWithDiscriminator("3", "bar", mapping.DiscriminatorValue, mapping),
			fixEventWithDiscriminator("4", "baz", mapping.DiscriminatorValue, mapping),
		}
		expected := []model.BusinessTenantMappingInput{
			{
				Name:           "foo",
				ExternalTenant: "1",
				Provider:       testProvider,
			},
			{
				Name:           "bar",
				ExternalTenant: "3",
				Provider:       testProvider,
			},
			{
				Name:           "baz",
				ExternalTenant: "4",
				Provider:       testProvider,
			},
		}

		// WHEN
		out := conv.EventsToTenants(eventsType, in)

		// THEN
		assert.Len(t, out, 3)
		assert.ElementsMatch(t, expected, out)
	})

	t.Run("Success when converting deleted events", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.DeletedEventsType
		in := []tenantfetcher.Event{
			fixEvent("1", "foo", testMapping),
			fixEvent("2", "bar", testMapping),
			fixEvent("3", "baz", testMapping),
		}
		expected := []model.BusinessTenantMappingInput{
			{
				Name:           "foo",
				ExternalTenant: "1",
				Provider:       testProvider,
			},
			{
				Name:           "bar",
				ExternalTenant: "2",
				Provider:       testProvider,
			},
			{
				Name:           "baz",
				ExternalTenant: "3",
				Provider:       testProvider,
			},
		}

		// WHEN
		out := conv.EventsToTenants(eventsType, in)

		// THEN
		assert.Len(t, out, len(in))
		assert.ElementsMatch(t, expected, out)
	})

	t.Run("Success when converting updated events", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.UpdatedEventsType
		in := []tenantfetcher.Event{
			fixEvent("1", "foo", testMapping),
			fixEvent("2", "bar", testMapping),
			fixEvent("3", "baz", testMapping),
		}
		expected := []model.BusinessTenantMappingInput{
			{
				Name:           "foo",
				ExternalTenant: "1",
				Provider:       testProvider,
			},
			{
				Name:           "bar",
				ExternalTenant: "2",
				Provider:       testProvider,
			},
			{
				Name:           "baz",
				ExternalTenant: "3",
				Provider:       testProvider,
			},
		}

		// WHEN
		out := conv.EventsToTenants(eventsType, in)

		// THEN
		assert.Len(t, out, len(in))
		assert.ElementsMatch(t, expected, out)
	})

	t.Run("Success when converting events and some are invalid", func(t *testing.T) {
		// GIVEN
		mapping := testMapping
		mapping.DiscriminatorField = testFieldDiscriminator
		mapping.DiscriminatorValue = testValueDiscriminator

		conv := tenantfetcher.NewConverter(testProvider, mapping)

		eventsType := tenantfetcher.CreatedEventsType
		in := []tenantfetcher.Event{
			fixEventWithDiscriminator("1", "foo", mapping.DiscriminatorValue, mapping),
			fixEventWithDiscriminator("2", "error", "ignore_me", mapping),
			fixEventWithDiscriminator("3", "bar", mapping.DiscriminatorValue, mapping),
			fixEventWithDiscriminator("4", "baz", mapping.DiscriminatorValue, mapping),
			tenantfetcher.Event{
				"badFormat": "bad",
			},
			tenantfetcher.Event{
				"eventData": fmt.Sprintf(`{"%s": 1, "%s": "123"}`, testMapping.IDField, testMapping.NameField),
			},
		}
		expected := []model.BusinessTenantMappingInput{
			{
				Name:           "foo",
				ExternalTenant: "1",
				Provider:       testProvider,
			},
			{
				Name:           "bar",
				ExternalTenant: "3",
				Provider:       testProvider,
			},
			{
				Name:           "baz",
				ExternalTenant: "4",
				Provider:       testProvider,
			},
		}

		// WHEN
		out := conv.EventsToTenants(eventsType, in)

		// THEN
		assert.Len(t, out, 3)
		assert.ElementsMatch(t, expected, out)
	})
}

func TestConverter_EventToTenant(t *testing.T) {
	t.Run("Error when no eventData", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.CreatedEventsType
		in := tenantfetcher.Event{
			"id":      "test",
			"invalid": "format",
		}

		// WHEN
		out, err := conv.EventToTenant(eventsType, in)

		// THEN
		require.EqualError(t, err, "invalid event data format")
		require.Empty(t, out)
	})

	t.Run("Error when eventData is not a valid json", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.CreatedEventsType
		in := tenantfetcher.Event{
			"id":        "test",
			"eventData": "{",
		}

		// WHEN
		out, err := conv.EventToTenant(eventsType, in)

		// THEN
		require.EqualError(t, err, "while unmarshalling event data: unexpected end of JSON input")
		require.Empty(t, out)
	})

	t.Run("Error when invalid format of discriminator field", func(t *testing.T) {
		// GIVEN
		mapping := testMapping
		mapping.DiscriminatorField = testFieldDiscriminator
		mapping.DiscriminatorValue = testValueDiscriminator

		conv := tenantfetcher.NewConverter(testProvider, mapping)
		eventsType := tenantfetcher.CreatedEventsType
		in := tenantfetcher.Event{
			"id":        "test",
			"eventData": fmt.Sprintf(`{"%s": "1", "%s": "test_name", "%s": 1}`, mapping.IDField, mapping.NameField, mapping.DiscriminatorField),
		}

		// WHEN
		out, err := conv.EventToTenant(eventsType, in)

		// THEN
		require.EqualError(t, err, "invalid format of discriminator field")
		require.Empty(t, out)
	})

	t.Run("Error when discriminator field missing when required", func(t *testing.T) {
		// GIVEN
		mapping := testMapping
		mapping.DiscriminatorField = testFieldDiscriminator
		mapping.DiscriminatorValue = testValueDiscriminator

		conv := tenantfetcher.NewConverter(testProvider, mapping)
		eventsType := tenantfetcher.CreatedEventsType
		in := tenantfetcher.Event{
			"id":        "test",
			"eventData": fmt.Sprintf(`{"%s": "1", "%s": "test_name"}`, mapping.IDField, mapping.NameField),
		}

		// WHEN
		out, err := conv.EventToTenant(eventsType, in)

		// THEN
		require.EqualError(t, err, "invalid format of discriminator field")
		require.Empty(t, out)
	})

	t.Run("Error when invalid format of id field", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.DeletedEventsType
		in := tenantfetcher.Event{
			"id":        "test",
			"eventData": fmt.Sprintf(`{"%s": 1, "%s": "123"}`, testMapping.IDField, testMapping.NameField),
		}

		// WHEN
		out, err := conv.EventToTenant(eventsType, in)

		// THEN
		require.EqualError(t, err, "invalid format of id field")
		require.Empty(t, out)
	})

	t.Run("Error when invalid format of name field", func(t *testing.T) {
		// GIVEN
		conv := tenantfetcher.NewConverter(testProvider, testMapping)
		eventsType := tenantfetcher.DeletedEventsType
		in := tenantfetcher.Event{
			"id":        "test",
			"eventData": fmt.Sprintf(`{"%s": "1", "%s": 123}`, testMapping.IDField, testMapping.NameField),
		}

		// WHEN
		out, err := conv.EventToTenant(eventsType, in)

		// THEN
		require.EqualError(t, err, "invalid format of name field")
		require.Empty(t, out)
	})
}
