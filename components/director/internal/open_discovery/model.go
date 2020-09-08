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
	Extensions       *string   `json:"extensions"` // TODO: Parse
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
	AssociatedPackages []string  `json:"associatedPackages"` // TODO: Parse
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
	Version          *string   `json:"version"`        // TODO: Parse
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
	AssociatedBundle string    `json:"associatedBundle"` // TODO: Parse
}

func (a *APIResource) ToAPIDefinitionInput() *graphql.APIDefinitionInput {
	return &graphql.APIDefinitionInput{
		ID:               &a.ID,
		Title:            a.Title,
		ShortDescription: a.ShortDescription,
		Description:      a.Description,
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
	Version          *string   `json:"version"`          // TODO: Parse
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
	AssociatedBundle string    `json:"associatedBundle"` // TODO: Parse
}

func (e *EventResource) ToEventDefinitionInput() *graphql.EventDefinitionInput {
	return &graphql.EventDefinitionInput{
		ID:               &e.ID,
		Title:            e.Title,
		ShortDescription: e.ShortDescription,
		Description:      e.Description,
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

//go:generate mockery -name=OpenDiscoveryDocumentConverter -output=automock -outpkg=automock -case=underscore
type OpenDiscoveryDocumentConverter interface {
	DocumentToGraphQLInputs(*Document) ([]*graphql.PackageInput, []*graphql.BundleInput, error)
}

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) DocumentToGraphQLInputs(in *Document) ([]*graphql.PackageInput, []*graphql.BundleInput, error) {
	if in == nil {
		return nil, nil, nil
	}
	pkgs := make([]*graphql.PackageInput, 0, len(in.Packages))
	for _, pkg := range in.Packages {
		pkgs = append(pkgs, pkg.ToPackageInput())
	}
	bundles := make(map[string]*graphql.BundleInput, len(in.Bundles))
	for _, bundle := range in.Bundles {
		bundles[bundle.ID] = bundle.ToBundleInput()
	}

	for _, api := range in.APIResources {
		bundle, ok := bundles[api.AssociatedBundle]
		if !ok {
			return nil, nil, fmt.Errorf("api resource with id: %s has unknown associated bundle with id: %s", api.ID, api.AssociatedBundle)
		}
		bundle.APIDefinitions = append(bundle.APIDefinitions, api.ToAPIDefinitionInput())
	}

	for _, event := range in.EventResources {
		bundle, ok := bundles[event.AssociatedBundle]
		if !ok {
			return nil, nil, fmt.Errorf("event resource with id: %s has unknown associated bundle with id: %s", event.ID, event.AssociatedBundle)
		}
		bundle.EventDefinitions = append(bundle.EventDefinitions, event.ToEventDefinitionInput())
	}

	bundlesSlice := make([]*graphql.BundleInput, 0, len(bundles))
	for _, bundle := range bundles {
		bundlesSlice = append(bundlesSlice, bundle)
	}

	return pkgs, bundlesSlice, nil
}

func strPtrToJSONPtr(in *string) *graphql.JSON {
	if in == nil {
		return nil
	}
	out := graphql.JSON(*in)
	return &out
}
