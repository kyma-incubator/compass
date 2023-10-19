package processor_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	ID1            = "ID1"
	ID2            = "ID2"
	ordID          = "com.compass.v1"
	localID        = "BusinessPartner"
	correlationIDs = `["sap.s4:sot:BusinessPartner", "sap.s4:sot:CostCenter", "sap.s4:sot:WorkforcePerson"]`
	level          = "aggregate"
	title          = "BusinessPartner"
	ordPackageID   = "sap.xref:package:SomePackage:v1"
	packageID      = "ppppppppp-pppp-pppp-pppp-pppppppppppp"
	vendorORDID    = "sap:vendor:SAP:"
	baseURL        = "http://test.com:8080"

	publicVisibility = "public"
	products         = `["sap:product:S4HANA_OD:"]`
	releaseStatus    = "active"
	entityTypeID     = "entity-type-id"
)

var (
	fixedTimestamp         = time.Now()
	appID                  = "appID"
	appTemplateVersionID   = "appTemplateVersionID"
	shortDescription       = "A business partner is a person, an organization, or a group of persons or organizations in which a company has a business interest."
	description            = "A workforce person is a natural person with a work agreement or relationship in form of a work assignment; it can be an employee or a contingent worker.\n"
	systemInstanceAware    = false
	policyLevel            = "custom"
	customPolicyLevel      = "sap:core:v1"
	sunsetDate             = "2022-01-08T15:47:04+00:00"
	successors             = `["sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1"]`
	extensible             = `{"supported":"automatic","description":"Please find the extensibility documentation"}`
	tags                   = `["storage","high-availability"]`
	resourceHash           = "123456"
	uint64ResourceHash     = uint64(123456)
	versionValue           = "v1.1"
	versionDeprecated      = false
	versionDeprecatedSince = "v1.0"
	versionForRemoval      = false
	changeLogEntries       = removeWhitespace(`[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`)
	links = removeWhitespace(`[
        {
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 1",
          "url": "https://example.com/2018/04/11/testing/"
        },
		{
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 2",
          "url": "http://example.com/testing/"
        }
      ]`)
	labels = removeWhitespace(`{
        "label-key-1": [
          "label-value-1",
          "label-value-2"
        ]
      }`)
	packageLabels = removeWhitespace(`{
        "label-key-1": [
          "label-val"
        ],
		"pkg-label": [
          "label-val"
        ]
      }`)
	packageLinksFormat = removeWhitespace(`[
        {
          "type": "terms-of-service",
          "url": "https://example.com/en/legal/terms-of-use.html"
        },
        {
          "type": "client-registration",
          "url": "%s/ui/public/showRegisterForm"
        }
      ]`)
	linksFormat = removeWhitespace(`[
        {
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 1",
          "url": "https://example.com/2018/04/11/testing/"
        },
		{
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 2",
          "url": "%s/testing/relative"
        }
      ]`)
	documentationLabels = removeWhitespace(`{
        "Some Aspect": ["Markdown Documentation [with links](#)", "With multiple values"]
      }`)
	errTest = errors.New("test error")
)

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\n", ""), "\t", "")
}

func fixVersionModel(value string, deprecated bool, deprecatedSince string, forRemoval bool) *model.Version {
	return &model.Version{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixEntityTypeModel(entityTypeID string) *model.EntityType {
	return &model.EntityType{
		BaseEntity: &model.BaseEntity{
			ID:        entityTypeID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
		ApplicationID:                &appID,
		ApplicationTemplateVersionID: &appTemplateVersionID,
		OrdID:                        ordID,
		LocalID:                      localID,
		CorrelationIDs:               json.RawMessage(correlationIDs),
		Level:                        level,
		Title:                        title,
		ShortDescription:             &shortDescription,
		Description:                  &description,
		SystemInstanceAware:          &systemInstanceAware,
		ChangeLogEntries:             json.RawMessage(changeLogEntries),
		PackageID:                    ordPackageID,
		Visibility:                   publicVisibility,
		Links:                        json.RawMessage(links),
		PartOfProducts:               json.RawMessage(products),
		PolicyLevel:                  &policyLevel,
		CustomPolicyLevel:            &customPolicyLevel,
		ReleaseStatus:                releaseStatus,
		SunsetDate:                   &sunsetDate,
		Successors:                   json.RawMessage(successors),
		Extensible:                   json.RawMessage(extensible),
		Tags:                         json.RawMessage(tags),
		Labels:                       json.RawMessage(labels),
		DocumentationLabels:          json.RawMessage(documentationLabels),
		Version:                      fixVersionModel(versionValue, versionDeprecated, versionDeprecatedSince, versionForRemoval),
		ResourceHash:                 &resourceHash,
	}
}

func fixEntityTypeInputModel() *model.EntityTypeInput {
	return &model.EntityTypeInput{
		OrdID:               ordID,
		LocalID:             localID,
		CorrelationIDs:      json.RawMessage(correlationIDs),
		Level:               level,
		Title:               title,
		ShortDescription:    &shortDescription,
		Description:         &description,
		SystemInstanceAware: &systemInstanceAware,
		ChangeLogEntries:    json.RawMessage(changeLogEntries),
		OrdPackageID:        ordPackageID,
		Visibility:          publicVisibility,
		Links:               json.RawMessage(links),
		PartOfProducts:      json.RawMessage(products),
		PolicyLevel:         &policyLevel,
		CustomPolicyLevel:   &customPolicyLevel,
		ReleaseStatus:       releaseStatus,
		SunsetDate:          &sunsetDate,
		Successors:          json.RawMessage(successors),
		Extensible:          json.RawMessage(extensible),
		Tags:                json.RawMessage(tags),
		Labels:              json.RawMessage(labels),
		DocumentationLabels: json.RawMessage(documentationLabels),
	}
}

func fixPackages() []*model.Package {
	return []*model.Package{
		{
			ID:                  packageID,
			ApplicationID:       &appID,
			OrdID:               ordPackageID,
			Vendor:              str.Ptr(vendorORDID),
			Title:               "PACKAGE 1 TITLE",
			ShortDescription:    "lorem ipsum",
			Description:         "lorem ipsum dolor set",
			Version:             "1.1.2",
			PackageLinks:        json.RawMessage(fmt.Sprintf(packageLinksFormat, baseURL)),
			Links:               json.RawMessage(fmt.Sprintf(linksFormat, baseURL)),
			LicenseType:         str.Ptr("licence"),
			SupportInfo:         str.Ptr("support-info"),
			Tags:                json.RawMessage(`["testTag"]`),
			Countries:           json.RawMessage(`["BG","EN"]`),
			Labels:              json.RawMessage(packageLabels),
			DocumentationLabels: json.RawMessage(documentationLabels),
			PolicyLevel:         str.Ptr(policyLevel),
			PartOfProducts:      json.RawMessage(products),
			LineOfBusiness:      json.RawMessage(`["Finance","Sales"]`),
			Industry:            json.RawMessage(`["Automotive","Banking","Chemicals"]`),
		},
	}
}
