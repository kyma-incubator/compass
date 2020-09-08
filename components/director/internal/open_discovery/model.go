package open_discovery

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	ShortDescription string    `json:"shortDescription"`
	Description      string    `json:"description"`
	Version          string    `json:"version"`
	Licence          *string   `json:"licence"`
	LicenceType      *string   `json:"licenceType"`
	TermsOfService   *string   `json:"termsOfService"`
	Logo             *string   `json:"logo"`
	Image            *string   `json:"image"`
	Provider         *string   `json:"provider"`
	Actions          *string   `json:"actions"`
	Tags             *string   `json:"tags"`
	LastUpdated      time.Time `json:"lastUpdated"`
	Extensions       *string   `json:"extensions"`
}

func (p *Package) ToPackageInput() *graphql.PackageInput {
	return &graphql.PackageInput{
		ID:               &p.ID,
		Title:            p.Title,
		ShortDescription: p.ShortDescription,
		Description:      p.Description,
		Version:          p.Version,
		Licence:          p.Licence,
		LicenceType:      p.LicenceType,
		TermsOfService:   p.TermsOfService,
		Logo:             p.Logo,
		Image:            p.Image,
		Provider:         strPtrToJSONPtr(p.Provider),
		Actions:          strPtrToJSONPtr(p.Actions),
		Tags:             strPtrToJSONPtr(p.Tags),
		LastUpdated:      graphql.Timestamp(p.LastUpdated),
		Extensions:       strPtrToJSONPtr(p.Extensions),
	}
}

type Bundle struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	ShortDescription   string    `json:"shortDescription"`
	Description        *string   `json:"description"`
	Tags               *string   `json:"tags"`
	LastUpdated        time.Time `json:"lastUpdated"`
	Extensions         *string   `json:"extensions"`
	AssociatedPackages []string  `json:"associatedPackages"`
}

func (b *Bundle) ToBundleInput() *graphql.BundleInput {
	return &graphql.BundleInput{
		ID:               &b.ID,
		Title:            b.Title,
		ShortDescription: b.ShortDescription,
		Description:      b.Description,
		Tags:             strPtrToJSONPtr(b.Tags),
		LastUpdated:      graphql.Timestamp(b.LastUpdated),
		Extensions:       strPtrToJSONPtr(b.Tags),
	}
}

type APIResource struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	ShortDescription string    `json:"shortDescription"`
	Description      *string   `json:"description"`
	EntryPoint       string    `json:"entryPoint"`
	Version          string    `json:"version"`
	APIDefinitions   string    `json:"apiDefinitions"` // TODO: Parse for spec
	Tags             *string   `json:"tags"`
	Documentation    *string   `json:"documentation"`
	ChangelogEntries *string   `json:"changelogEntries"`
	Logo             *string   `json:"logo"`
	Image            *string   `json:"image"`
	URL              *string   `json:"url"`
	ReleaseStatus    string    `json:"releaseStatus"`
	APIProtocol      string    `json:"apiProtocol"`
	Actions          string    `json:"actions"`
	LastUpdated      time.Time `json:"lastUpdated"`
	Extensions       *string   `json:"extensions"`
	AssociatedBundle string    `json:"associatedBundle"`
}

func (a *APIResource) ToAPIDefinitionInput() *graphql.APIDefinitionInput {
	return &graphql.APIDefinitionInput{
		ID:               &a.ID,
		Title:            a.Title,
		ShortDescription: a.ShortDescription,
		Description:      a.Description,
		Version: &graphql.VersionInput{
			Value: a.Version,
		},
		EntryPoint:       a.EntryPoint,
		APIDefinitions:   graphql.JSON(a.APIDefinitions),
		Tags:             strPtrToJSONPtr(a.Tags),
		Documentation:    a.Documentation,
		ChangelogEntries: strPtrToJSONPtr(a.ChangelogEntries),
		Logo:             a.Logo,
		Image:            a.Image,
		URL:              a.URL,
		ReleaseStatus:    a.ReleaseStatus,
		APIProtocol:      a.APIProtocol,
		Actions:          graphql.JSON(a.Actions),
		LastUpdated:      graphql.Timestamp(a.LastUpdated),
		Extensions:       strPtrToJSONPtr(a.Extensions),
	}
}

type EventResource struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	ShortDescription string    `json:"shortDescription"`
	Description      *string   `json:"description"`
	Version          string    `json:"version"`
	EventDefinitions string    `json:"eventDefinitions"` // TODO: Parse
	Tags             *string   `json:"tags"`
	Documentation    *string   `json:"documentation"`
	ChangelogEntries *string   `json:"changelogEntries"`
	Logo             *string   `json:"logo"`
	Image            *string   `json:"image"`
	URL              *string   `json:"url"`
	ReleaseStatus    string    `json:"releaseStatus"`
	LastUpdated      time.Time `json:"lastUpdated"`
	Extensions       *string   `json:"extensions"`
	AssociatedBundle string    `json:"associatedBundle"`
}

func (e *EventResource) ToEventDefinitionInput() *graphql.EventDefinitionInput {
	return &graphql.EventDefinitionInput{
		ID:               &e.ID,
		Title:            e.Title,
		ShortDescription: e.ShortDescription,
		Description:      e.Description,
		Version: &graphql.VersionInput{
			Value: e.Version,
		},
		EventDefinitions: graphql.JSON(e.EventDefinitions),
		Tags:             strPtrToJSONPtr(e.Tags),
		Documentation:    e.Documentation,
		ChangelogEntries: strPtrToJSONPtr(e.ChangelogEntries),
		Logo:             e.Logo,
		Image:            e.Image,
		URL:              e.URL,
		ReleaseStatus:    e.ReleaseStatus,
		LastUpdated:      graphql.Timestamp(e.LastUpdated),
		Extensions:       strPtrToJSONPtr(e.Extensions),
	}
}

type Document struct {
	Schema               string           `json:"$schema"`
	OpenDiscoveryVersion string           `json:"openDiscovery"`
	BaseURL              string           `json:"baseUrl"`
	LastUpdated          time.Time        `json:"lastUpdated"`
	Extensions           *string          `json:"extensions"`
	Packages             []*Package       `json:"packages"`
	Bundles              []*Bundle        `json:"bundles"`
	APIResources         []*APIResource   `json:"apiResources"`
	EventResources       []*EventResource `json:"eventResources"`
}

func (d *Document) ToGraphQLInputs() ([]*graphql.PackageInput, map[string]*graphql.BundleInput, error) {
	if d == nil {
		return nil, nil, nil
	}
	pkgs := make([]*graphql.PackageInput, 0, len(d.Packages))
	for _, pkg := range d.Packages {
		pkgs = append(pkgs, pkg.ToPackageInput())
	}
	bundles := make(map[string]*graphql.BundleInput, len(d.Bundles))
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

	resultBundles := make(map[string]*graphql.BundleInput, 0)
	for _, bundle := range d.Bundles {
		for _, pkgID := range bundle.AssociatedPackages {
			resultBundles[pkgID] = bundles[bundle.ID]
		}
	}

	return pkgs, resultBundles, nil
}

func strPtrToJSONPtr(in *string) *graphql.JSON {
	if in == nil {
		return nil
	}
	out := graphql.JSON(*in)
	return &out
}
