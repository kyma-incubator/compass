package nsmodel

import (
	"fmt"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/require"
)

type cases []struct {
	data          validation.Validatable
	name          string
	expectErr     bool
	expectedError string
}

func TestReport_Validate(t *testing.T) {
	validation.ErrRequired = validation.ErrRequired.SetMessage("the value is required")
	validation.ErrNotNilRequired = validation.ErrNotNilRequired.SetMessage("the value can not be nil")

	validSystem := System{
		SystemBase: SystemBase{
			Protocol:     "protocol",
			Host:         "host",
			SystemType:   "systemType",
			Description:  "description",
			Status:       "status",
			SystemNumber: "systemNumber",
		},
		TemplateID: "emptyTemplateID",
	}

	validSCC := SCC{
		ExternalSubaccountID: "external-subaccount",
		InternalSubaccountID: "",
		LocationID:           "loc-id",
		ExposedSystems:       []System{validSystem},
	}

	invalidSCC := SCC{
		ExternalSubaccountID: "",
		InternalSubaccountID: "",
		LocationID:           "loc-id",
		ExposedSystems:       []System{},
	}

	cases := cases{
		{
			data: Report{
				ReportType: "delta",
				Value:      []SCC{validSCC},
			},
			name:      "success",
			expectErr: false,
		},
		{
			data: Report{
				ReportType: "delta",
				Value:      []SCC{},
			},
			name:      "success with empty slice of SCCs",
			expectErr: false,
		},
		{
			data: Report{
				Value: []SCC{validSCC},
			},
			name:          "fail with missing report type",
			expectErr:     true,
			expectedError: "type: the value is required.",
		},
		{
			data: Report{
				Value: []SCC{},
			},
			name:          "fail with empty report type",
			expectErr:     true,
			expectedError: "type: the value is required.",
		},
		{
			data: Report{
				ReportType: "delta",
			},
			name:          "fail with missing value",
			expectErr:     true,
			expectedError: "value: the value can not be nil.",
		},
		{
			data: Report{
				ReportType: "delta",
				Value:      []SCC{invalidSCC},
			},
			name:          "fail with invalid SCC",
			expectErr:     true,
			expectedError: "value: (subaccount: the value is required.).",
		},
	}
	checkCases(cases, t)
}

func TestSCC_Validate(t *testing.T) {
	validation.ErrRequired = validation.ErrRequired.SetMessage("the value is required")
	validation.ErrNotNilRequired = validation.ErrNotNilRequired.SetMessage("the value can not be nil")

	validSystem := System{
		SystemBase: SystemBase{
			Protocol:     "protocol",
			Host:         "host",
			SystemType:   "systemType",
			Description:  "description",
			Status:       "status",
			SystemNumber: "systemNumber",
		},
		TemplateID: "emptyTemplateID",
	}

	invalidSystem := System{
		SystemBase: SystemBase{
			Host:         "host",
			SystemType:   "systemType",
			Description:  "description",
			Status:       "status",
			SystemNumber: "systemNumber",
		},
		TemplateID: "emptyTemplateID",
	}

	cases := cases{
		{
			data: SCC{
				ExternalSubaccountID: "external-subaccount",
				InternalSubaccountID: "",
				LocationID:           "loc-id",
				ExposedSystems:       []System{validSystem},
			},
			name:      "success",
			expectErr: false,
		},
		{
			data: SCC{
				ExternalSubaccountID: "external-subaccount",
				InternalSubaccountID: "",
				LocationID:           "",
				ExposedSystems:       []System{validSystem},
			},
			name:      "success with empty location ID",
			expectErr: false,
		},
		{
			data: SCC{
				ExternalSubaccountID: "external-subaccount",
				InternalSubaccountID: "",
				LocationID:           "loc-id",
				ExposedSystems:       []System{},
			},
			name:      "success with empty slice of systems",
			expectErr: false,
		},
		{
			data: SCC{
				ExternalSubaccountID: "external-subaccount",
				LocationID:           "loc-id",
				ExposedSystems:       []System{},
			},
			name:      "success with missing internal subaccount id",
			expectErr: false,
		},
		{
			data: SCC{
				ExternalSubaccountID: "",
				InternalSubaccountID: "",
				LocationID:           "loc-id",
				ExposedSystems:       []System{},
			},
			name:          "fail with empty subaccount",
			expectErr:     true,
			expectedError: "subaccount: the value is required.",
		},
		{
			data: SCC{
				LocationID:     "loc-id",
				ExposedSystems: []System{},
			},
			name:          "fail with missing subaccount",
			expectErr:     true,
			expectedError: "subaccount: the value is required.",
		},
		{
			data: SCC{
				ExternalSubaccountID: "external-subaccount",
				InternalSubaccountID: "",
				LocationID:           "loc-id",
			},
			name:          "fail with missing ExposedSystem",
			expectErr:     true,
			expectedError: "exposedSystems: the value can not be nil.",
		},
		{
			data: SCC{
				ExternalSubaccountID: "external-subaccount",
				InternalSubaccountID: "",
				LocationID:           "loc-id",
				ExposedSystems:       []System{invalidSystem},
			},
			name:          "fail with invalid system",
			expectErr:     true,
			expectedError: "exposedSystems: (protocol: the value is required.).",
		},
	}

	checkCases(cases, t)
}

func TestSystem_Validate(t *testing.T) {
	validation.ErrRequired = validation.ErrRequired.SetMessage("the value is required")
	validation.ErrNotNilRequired = validation.ErrNotNilRequired.SetMessage("the value can not be nil")

	protocol := "HTTP"
	host := "127.0.0.1:8080"
	systemType := "nonSAPsys"
	description := "description"
	status := "unreachable"
	systemNumber := "sys-num"
	emptyTemplateID := ""

	cases := cases{
		{
			name: "success with all fields present",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         host,
					SystemType:   systemType,
					Description:  description,
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr: false,
		},
		{
			name: "success with empty description",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         host,
					SystemType:   systemType,
					Description:  "",
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr: false,
		},
		{
			name: "success with empty systemNumber",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         host,
					SystemType:   systemType,
					Description:  description,
					Status:       status,
					SystemNumber: "",
				},
				TemplateID: emptyTemplateID,
			},
			expectErr: false,
		},
		{
			name: "fail when missing protocol",
			data: System{
				SystemBase: SystemBase{
					Host:         host,
					SystemType:   systemType,
					Description:  description,
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "protocol: the value is required.",
		},
		{
			name: "fail when protocol is empty",
			data: System{
				SystemBase: SystemBase{
					Protocol:     "",
					Host:         host,
					SystemType:   systemType,
					Description:  description,
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "protocol: the value is required.",
		},
		{
			name: "fail when missing host",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					SystemType:   systemType,
					Description:  description,
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "host: the value is required.",
		},
		{
			name: "fail when host is empty",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         "",
					SystemType:   systemType,
					Description:  description,
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "host: the value is required.",
		},
		{
			name: "fail when missing systemType",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         host,
					Description:  description,
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "type: the value is required.",
		},
		{
			name: "fail when systemType is empty",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         host,
					SystemType:   "",
					Description:  description,
					Status:       status,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "type: the value is required.",
		},
		{
			name: "fail when missing status",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         host,
					SystemType:   systemType,
					Description:  description,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "status: the value is required.",
		},
		{
			name: "fail when status is empty",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					Host:         host,
					SystemType:   systemType,
					Description:  description,
					Status:       "",
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "status: the value is required.",
		},
		{
			name: "fail when multiple fields are missing",
			data: System{
				SystemBase: SystemBase{
					Protocol:     protocol,
					SystemType:   systemType,
					Description:  description,
					SystemNumber: systemNumber,
				},
				TemplateID: emptyTemplateID,
			},
			expectErr:     true,
			expectedError: "host: the value is required; status: the value is required.",
		},
	}

	checkCases(cases, t)
}

func TestSystem_UnmarshalJSON(t *testing.T) {
	t.Run("fail to marshal system", func(t *testing.T) {
		s := &System{
			SystemBase: SystemBase{},
			TemplateID: "",
		}
		err := s.UnmarshalJSON(nil)

		require.Error(t, err)
	})

	t.Run("success when template is matched", func(t *testing.T) {
		systemString := "{\"protocol\": \"HTTP\",\"host\": \"127.0.0.1:8080\",\"type\": \"otherSAPsys\",\"status\": \"disabled\",\"description\": \"description\"}"

		Mappings = append(Mappings,
			TemplateMapping{
				Name:        "",
				ID:          "sss",
				SourceKey:   []string{"type"},
				SourceValue: []string{"type"},
			},
			TemplateMapping{
				Name:        "",
				ID:          "ss",
				SourceKey:   []string{"description"},
				SourceValue: []string{"description"},
			})

		actualSystem := &System{
			SystemBase: SystemBase{},
			TemplateID: "",
		}
		expectedSystem := &System{
			SystemBase: SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "",
			},
			TemplateID: "ss",
		}

		err := actualSystem.UnmarshalJSON([]byte(systemString))
		require.NoError(t, err)
		require.Equal(t, expectedSystem, actualSystem)

		Mappings = nil
	})

	t.Run("success when do not match to any template", func(t *testing.T) {
		systemString := "{\"protocol\": \"HTTP\",\"host\": \"127.0.0.1:8080\",\"type\": \"otherSAPsys\",\"status\": \"disabled\",\"description\": \"description\"}"

		Mappings = append(Mappings,
			TemplateMapping{
				Name:        "",
				ID:          "sss",
				SourceKey:   []string{"type"},
				SourceValue: []string{"type"},
			},
			TemplateMapping{
				Name:        "",
				ID:          "ss",
				SourceKey:   []string{"description"},
				SourceValue: []string{"sth"},
			})

		actualSystem := &System{
			SystemBase: SystemBase{},
			TemplateID: "",
		}
		expectedSystem := &System{
			SystemBase: SystemBase{
				Protocol:     "HTTP",
				Host:         "127.0.0.1:8080",
				SystemType:   "otherSAPsys",
				Description:  "description",
				Status:       "disabled",
				SystemNumber: "",
			},
			TemplateID: "",
		}

		err := actualSystem.UnmarshalJSON([]byte(systemString))
		require.NoError(t, err)
		require.Equal(t, expectedSystem, actualSystem)

		Mappings = nil
	})
}

func checkCases(cases cases, t *testing.T) {
	for _, c := range cases {
		t.Run(fmt.Sprintf("Checking case %s", c.name), func(t *testing.T) {
			err := c.data.Validate()

			if c.expectErr {
				require.Equal(t, c.expectedError, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
