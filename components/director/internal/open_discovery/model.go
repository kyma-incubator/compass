package open_discovery

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"time"
)

type WellKnownConfig struct {
	Schema                string                `json:"$schema"`
	OpenDiscoveryV1Config OpenDiscoveryV1Config `json:"open-discovery-v1"`
}

type OpenDiscoveryV1Config struct {
	DocumentConfig DocumentConfig `json:"document"`
}

type DocumentConfig struct {
	URL string `json:"url"`
}

type Package struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	ShortDescription string          `json:"shortDescription"`
	Description      string          `json:"description"`
	Version          string          `json:"version"`
	Licence          *string         `json:"licence"`
	LicenceType      *string         `json:"licenceType"`
	TermsOfService   *string         `json:"termsOfService"`
	Logo             *string         `json:"logo"`
	Image            *string         `json:"image"`
	Provider         json.RawMessage `json:"provider"`
	Actions          json.RawMessage `json:"actions"`
	Tags             json.RawMessage `json:"tags"`
	LastUpdated      time.Time       `json:"lastUpdated"`
	Extensions       json.RawMessage `json:"extensions"`
}

func (p *Package) ToPackageInput() *model.PackageInput {
	return &model.PackageInput{
		ID:               p.ID,
		Title:            p.Title,
		ShortDescription: p.ShortDescription,
		Description:      p.Description,
		Version:          p.Version,
		Licence:          p.Licence,
		LicenceType:      p.LicenceType,
		TermsOfService:   p.TermsOfService,
		Logo:             p.Logo,
		Image:            p.Image,
		Provider:         rawJsonToStrPtr(p.Provider),
		Actions:          rawJsonToStrPtr(p.Actions),
		Tags:             rawJsonToStrPtr(p.Tags),
		LastUpdated:      p.LastUpdated,
		Extensions:       rawJsonToStrPtr(p.Extensions),
	}
}

type Bundle struct {
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	ShortDescription   string          `json:"shortDescription"`
	Description        *string         `json:"description"`
	Tags               json.RawMessage `json:"tags"`
	LastUpdated        time.Time       `json:"lastUpdated"`
	Extensions         json.RawMessage `json:"extensions"`
	AssociatedPackages []string        `json:"associatedPackages"`
}

func (b *Bundle) ToBundleInput() *model.BundleInput {
	return &model.BundleInput{
		ID:               b.ID,
		Title:            b.Title,
		ShortDescription: b.ShortDescription,
		Description:      b.Description,
		Tags:             rawJsonToStrPtr(b.Tags),
		LastUpdated:      b.LastUpdated,
		Extensions:       rawJsonToStrPtr(b.Extensions),
	}
}

type APIResource struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	ShortDescription string          `json:"shortDescription"`
	Description      *string         `json:"description"`
	EntryPoint       string          `json:"entryPoint"`
	Version          string          `json:"version"`
	APIDefinitions   json.RawMessage `json:"apiDefinitions"` // TODO: Parse for spec
	Tags             json.RawMessage `json:"tags"`
	Documentation    *string         `json:"documentation"`
	ChangelogEntries json.RawMessage `json:"changelogEntries"`
	Logo             *string         `json:"logo"`
	Image            *string         `json:"image"`
	URL              *string         `json:"url"`
	ReleaseStatus    string          `json:"releaseStatus"`
	APIProtocol      string          `json:"apiProtocol"`
	Actions          json.RawMessage `json:"actions"`
	LastUpdated      time.Time       `json:"lastUpdated"`
	Extensions       json.RawMessage `json:"extensions"`
	AssociatedBundle string          `json:"associatedBundle"`
}

func (a *APIResource) ToAPIDefinitionInput() *model.APIDefinitionInput {
	return &model.APIDefinitionInput{
		ID:               a.ID,
		Title:            a.Title,
		ShortDescription: a.ShortDescription,
		Description:      a.Description,
		Version: &model.VersionInput{
			Value: a.Version,
		},
		EntryPoint:       a.EntryPoint,
		APIDefinitions:   rawJsonToStr(a.APIDefinitions),
		Tags:             rawJsonToStrPtr(a.Tags),
		Documentation:    a.Documentation,
		ChangelogEntries: rawJsonToStrPtr(a.ChangelogEntries),
		Logo:             a.Logo,
		Image:            a.Image,
		URL:              a.URL,
		ReleaseStatus:    a.ReleaseStatus,
		APIProtocol:      a.APIProtocol,
		Actions:          rawJsonToStr(a.Actions),
		LastUpdated:      a.LastUpdated,
		Extensions:       rawJsonToStrPtr(a.Extensions),
	}
}

type EventResource struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	ShortDescription string          `json:"shortDescription"`
	Description      *string         `json:"description"`
	Version          string          `json:"version"`
	EventDefinitions json.RawMessage `json:"eventDefinitions"` // TODO: Parse
	Tags             json.RawMessage `json:"tags"`
	Documentation    *string         `json:"documentation"`
	ChangelogEntries json.RawMessage `json:"changelogEntries"`
	Logo             *string         `json:"logo"`
	Image            *string         `json:"image"`
	URL              *string         `json:"url"`
	ReleaseStatus    string          `json:"releaseStatus"`
	LastUpdated      time.Time       `json:"lastUpdated"`
	Extensions       json.RawMessage `json:"extensions"`
	AssociatedBundle string          `json:"associatedBundle"`
}

func (e *EventResource) ToEventDefinitionInput() *model.EventDefinitionInput {
	return &model.EventDefinitionInput{
		ID:               e.ID,
		Title:            e.Title,
		ShortDescription: e.ShortDescription,
		Description:      e.Description,
		Version: &model.VersionInput{
			Value: e.Version,
		},
		EventDefinitions: rawJsonToStr(e.EventDefinitions),
		Tags:             rawJsonToStrPtr(e.Tags),
		Documentation:    e.Documentation,
		ChangelogEntries: rawJsonToStrPtr(e.ChangelogEntries),
		Logo:             e.Logo,
		Image:            e.Image,
		URL:              e.URL,
		ReleaseStatus:    e.ReleaseStatus,
		LastUpdated:      e.LastUpdated,
		Extensions:       rawJsonToStrPtr(e.Extensions),
	}
}

type Document struct {
	Schema               string           `json:"$schema"`
	OpenDiscoveryVersion string           `json:"openDiscovery"`
	BaseURL              string           `json:"baseUrl"`
	LastUpdated          time.Time        `json:"lastUpdated"`
	Extensions           json.RawMessage  `json:"extensions"`
	Packages             []*Package       `json:"packages"`
	Bundles              []*Bundle        `json:"bundles"`
	APIResources         []*APIResource   `json:"apiResources"`
	EventResources       []*EventResource `json:"eventResources"`
}

type BundleInputWithAssociatedPackages struct {
	In                 *model.BundleInput
	AssociatedPackages []string
}

func (d *Document) ToModelInputs() ([]*model.PackageInput, []*BundleInputWithAssociatedPackages, error) {
	if d == nil {
		return nil, nil, nil
	}
	pkgs := make([]*model.PackageInput, 0, len(d.Packages))
	for _, pkg := range d.Packages {
		pkgs = append(pkgs, pkg.ToPackageInput())
	}
	bundles := make(map[string]*model.BundleInput, len(d.Bundles))
	for _, bundle := range d.Bundles {
		bundles[bundle.ID] = bundle.ToBundleInput()
	}

	for _, api := range d.APIResources {
		bundle, ok := bundles[api.AssociatedBundle]
		if !ok {
			return nil, nil, fmt.Errorf("api resource with id: %s has unknown associated bundle with id: %s", api.ID, api.AssociatedBundle)
		}
		bundle.APIDefinitions = append(bundle.APIDefinitions, api.ToAPIDefinitionInput())
	}

	for _, event := range d.EventResources {
		bundle, ok := bundles[event.AssociatedBundle]
		if !ok {
			return nil, nil, fmt.Errorf("event resource with id: %s has unknown associated bundle with id: %s", event.ID, event.AssociatedBundle)
		}
		bundle.EventDefinitions = append(bundle.EventDefinitions, event.ToEventDefinitionInput())
	}

	resultBundles := make([]*BundleInputWithAssociatedPackages, 0)
	for _, bundle := range d.Bundles {
		resultBundles = append(resultBundles, &BundleInputWithAssociatedPackages{
			In:                 bundles[bundle.ID],
			AssociatedPackages: bundle.AssociatedPackages,
		})
	}

	return pkgs, resultBundles, nil
}

func rawJsonToStrPtr(j json.RawMessage) *string {
	if j == nil {
		return nil
	}
	jstr := string(j)
	return &jstr
}

func rawJsonToStr(j json.RawMessage) string {
	if j == nil {
		return ""
	}
	return string(j)
}
