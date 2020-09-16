package open_discovery

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"net/url"
	"strings"
	"time"
)

type WellKnownConfig struct {
	Schema                string                `json:"$schema"`
	OpenDiscoveryV1Config OpenDiscoveryV1Config `json:"open-discovery-v1"`
}

type OpenDiscoveryV1Config struct {
	DocumentConfigs []DocumentConfig `json:"documents"`
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

func (p *Package) ToPackageInput(baseURL string) (*model.PackageInput, error) {
	baseURL, err := normalizeURL(baseURL)
	if err != nil {
		return nil, err
	}

	if p.TermsOfService != nil && !IsUrl(*p.TermsOfService) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("termsOfService for package with ID %s should be absolute URL if baseURL not provided in the document", p.ID)
		} else {
			*p.TermsOfService = baseURL + *p.TermsOfService
		}
	}
	if p.Licence != nil && !IsUrl(*p.Licence) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("license for package with ID %s should be absolute URL if baseURL not provided in the document", p.ID)
		} else {
			*p.Licence = baseURL + *p.Licence
		}
	}
	if p.Logo != nil && !IsUrl(*p.Logo) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("logo for package with ID %s should be absolute URL if baseURL not provided in the document", p.ID)
		} else {
			*p.Logo = baseURL + *p.Logo
		}
	}
	if p.Image != nil && !IsUrl(*p.Image) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("image for package with ID %s should be absolute URL if baseURL not provided in the document", p.ID)
		} else {
			*p.Image = baseURL + *p.Image
		}
	}
	if p.Provider, err = rewriteRelativeURLsInJson(p.Provider, baseURL, "logo"); err != nil {
		return nil, fmt.Errorf("error rewriting urls in provider for package with ID %s", p.ID)
	}
	if p.Actions, err = rewriteRelativeURLsInJson(p.Actions, baseURL, "target"); err != nil {
		return nil, fmt.Errorf("error rewriting urls in actions for package with ID %s", p.ID)
	}

	return &model.PackageInput{
		OpenDiscoveryID:  p.ID,
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
	}, nil
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
		OpenDiscoveryID:  b.ID,
		Title:            b.Title,
		ShortDescription: b.ShortDescription,
		Description:      b.Description,
		Tags:             rawJsonToStrPtr(b.Tags),
		LastUpdated:      b.LastUpdated,
		Extensions:       rawJsonToStrPtr(b.Extensions),
	}
}

type Spec struct {
	Type          string  `json:"type"`
	CustomType    *string `json:"customType"`
	URL           string  `json:"url"`
	Serialization string  `json:"serialization"`
}

func (s *Spec) ToAPISpecInput() *model.APISpecInput {
	return &model.APISpecInput{
		Type:       model.APISpecType(s.Type), // TODO: Validation
		CustomType: s.CustomType,
		Format:     model.SpecFormat(s.Serialization), // TODO: Validation
		FetchRequest: &model.FetchRequestInput{
			URL: s.URL,
		},
	}
}

func (s *Spec) ToEventSpecInput() *model.EventSpecInput {
	return &model.EventSpecInput{
		EventSpecType: model.EventSpecType(s.Type), // TODO: Validation
		CustomType:    s.CustomType,
		Format:        model.SpecFormat(s.Serialization), // TODO: Validation
		FetchRequest: &model.FetchRequestInput{
			URL: s.URL,
		},
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
	APISpecs         []Spec          `json:"-"`
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

func (a *APIResource) ToAPIDefinitionInput(baseURL string) (*model.APIDefinitionInput, error) {
	baseURL, err := normalizeURL(baseURL)
	if err != nil {
		return nil, err
	}

	if !IsUrl(a.EntryPoint) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("entryPoint for apiResource with ID %s should be absolute URL if baseURL not provided in the document", a.ID)
		} else {
			a.EntryPoint = baseURL + a.EntryPoint
		}
	}
	if a.Documentation != nil && !IsUrl(*a.Documentation) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("documentation for apiResource with ID %s should be absolute URL if baseURL not provided in the document", a.ID)
		} else {
			*a.Documentation = baseURL + *a.Documentation
		}
	}
	if a.URL != nil && !IsUrl(*a.URL) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("url for apiResource with ID %s should be absolute URL if baseURL not provided in the document", a.ID)
		} else {
			*a.URL = baseURL + *a.URL
		}
	}
	if a.Logo != nil && !IsUrl(*a.Logo) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("logo for apiResource with ID %s should be absolute URL if baseURL not provided in the document", a.ID)
		} else {
			*a.Logo = baseURL + *a.Logo
		}
	}
	if a.Image != nil && !IsUrl(*a.Image) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("image for apiResource with ID %s should be absolute URL if baseURL not provided in the document", a.ID)
		} else {
			*a.Image = baseURL + *a.Image
		}
	}
	if a.Actions, err = rewriteRelativeURLsInJson(a.Actions, baseURL, "target"); err != nil {
		return nil, fmt.Errorf("error rewriting urls in actions for apiResource with ID %s", a.ID)
	}
	if a.APIDefinitions, err = rewriteRelativeURLsInJson(a.APIDefinitions, baseURL, "url"); err != nil {
		return nil, fmt.Errorf("error rewriting urls in eventDefinitions for apiResource with ID %s", a.ID)
	}
	if a.ChangelogEntries, err = rewriteRelativeURLsInJson(a.ChangelogEntries, baseURL, "url"); err != nil {
		return nil, fmt.Errorf("error rewriting urls in changelogEntrie for apiResource with ID %s", a.ID)
	}

	if err := json.Unmarshal(a.APIDefinitions, &a.APISpecs); err != nil {
		return nil, err
	}

	specs := make([]*model.APISpecInput, 0, len(a.APISpecs))
	for _, spec := range a.APISpecs {
		specs = append(specs, spec.ToAPISpecInput())
	}

	return &model.APIDefinitionInput{
		OpenDiscoveryID:  a.ID,
		Title:            a.Title,
		ShortDescription: a.ShortDescription,
		Description:      a.Description,
		Version: &model.VersionInput{
			Value: a.Version,
		},
		Specs:            specs,
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
	}, nil
}

type EventResource struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	ShortDescription string          `json:"shortDescription"`
	Description      *string         `json:"description"`
	Version          string          `json:"version"`
	EventDefinitions json.RawMessage `json:"eventDefinitions"` // TODO: Parse for spec
	EventSpecs       []Spec          `json:"-"`
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

func (e *EventResource) ToEventDefinitionInput(baseURL string) (*model.EventDefinitionInput, error) {
	baseURL, err := normalizeURL(baseURL)
	if err != nil {
		return nil, err
	}

	if e.Documentation != nil && !IsUrl(*e.Documentation) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("documentation for eventResource with ID %s should be absolute URL if baseURL not provided in the document", e.ID)
		} else {
			*e.Documentation = baseURL + *e.Documentation
		}
	}
	if e.URL != nil && !IsUrl(*e.URL) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("url for eventResource with ID %s should be absolute URL if baseURL not provided in the document", e.ID)
		} else {
			*e.URL = baseURL + *e.URL
		}
	}
	if e.Logo != nil && !IsUrl(*e.Logo) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("logo for eventResource with ID %s should be absolute URL if baseURL not provided in the document", e.ID)
		} else {
			*e.Logo = baseURL + *e.Logo
		}
	}
	if e.Image != nil && !IsUrl(*e.Image) {
		if len(baseURL) == 0 {
			return nil, fmt.Errorf("image for eventResource with ID %s should be absolute URL if baseURL not provided in the document", e.ID)
		} else {
			*e.Image = baseURL + *e.Image
		}
	}
	if e.EventDefinitions, err = rewriteRelativeURLsInJson(e.EventDefinitions, baseURL, "url"); err != nil {
		return nil, fmt.Errorf("error rewriting urls in eventDefinitions for eventResource with ID %s", e.ID)
	}
	if e.ChangelogEntries, err = rewriteRelativeURLsInJson(e.ChangelogEntries, baseURL, "url"); err != nil {
		return nil, fmt.Errorf("error rewriting urls in changelogEntries for eventResource with ID %s", e.ID)
	}

	if err := json.Unmarshal(e.EventDefinitions, &e.EventSpecs); err != nil {
		return nil, err
	}

	specs := make([]*model.EventSpecInput, 0, len(e.EventSpecs))
	for _, spec := range e.EventSpecs {
		specs = append(specs, spec.ToEventSpecInput())
	}

	return &model.EventDefinitionInput{
		OpenDiscoveryID:  e.ID,
		Title:            e.Title,
		ShortDescription: e.ShortDescription,
		Description:      e.Description,
		Version: &model.VersionInput{
			Value: e.Version,
		},
		Specs:            specs,
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
	}, nil
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

type Documents []Document

type BundleInputWithAssociatedPackages struct {
	In                 *model.BundleInput
	AssociatedPackages []string
}

type ApiInputWithAssociatedBundle struct {
	In               *model.APIDefinitionInput
	AssociatedBundle string
}

type EventInputWithAssociatedBundle struct {
	In               *model.EventDefinitionInput
	AssociatedBundle string
}

func (docs Documents) ToModelInputs() (map[string]*model.PackageInput, []*BundleInputWithAssociatedPackages, error) {
	if docs == nil {
		return nil, nil, nil
	}
	pkgs := make(map[string]*model.PackageInput, 0)
	bundles := make(map[string]*BundleInputWithAssociatedPackages, 0)
	apis := make(map[string]*ApiInputWithAssociatedBundle)
	events := make(map[string]*EventInputWithAssociatedBundle)

	for _, d := range docs {
		for _, pkg := range d.Packages {
			if _, ok := pkgs[pkg.ID]; ok {
				return nil, nil, fmt.Errorf("package with id %s found in multiple documents", pkg.ID)
			}
			pkgInput, err := pkg.ToPackageInput(d.BaseURL)
			if err != nil {
				return nil, nil, err
			}
			pkgs[pkg.ID] = pkgInput
		}

		for _, bundle := range d.Bundles {
			if _, ok := bundles[bundle.ID]; ok {
				return nil, nil, fmt.Errorf("bundle with id %s found in multiple documents", bundle.ID)
			}
			bundles[bundle.ID] = &BundleInputWithAssociatedPackages{
				In:                 bundle.ToBundleInput(),
				AssociatedPackages: bundle.AssociatedPackages,
			}
		}

		for _, api := range d.APIResources {
			if _, ok := apis[api.ID]; ok {
				return nil, nil, fmt.Errorf("api with id %s found in multiple documents", api.ID)
			}
			apiInput, err := api.ToAPIDefinitionInput(d.BaseURL)
			if err != nil {
				return nil, nil, err
			}
			apis[api.ID] = &ApiInputWithAssociatedBundle{
				In:               apiInput,
				AssociatedBundle: api.AssociatedBundle,
			}
		}

		for _, event := range d.EventResources {
			if _, ok := events[event.ID]; ok {
				return nil, nil, fmt.Errorf("event with id %s found in multiple documents", event.ID)
			}
			eventInput, err := event.ToEventDefinitionInput(d.BaseURL)
			if err != nil {
				return nil, nil, err
			}
			events[event.ID] = &EventInputWithAssociatedBundle{
				In:               eventInput,
				AssociatedBundle: event.AssociatedBundle,
			}
		}
	}

	for _, api := range apis {
		bundle, ok := bundles[api.AssociatedBundle]
		if !ok {
			return nil, nil, fmt.Errorf("api resource with id: %s has unknown associated bundle with id: %s", api.In.ID, api.AssociatedBundle)
		}
		bundle.In.APIDefinitions = append(bundle.In.APIDefinitions, api.In)
	}

	for _, event := range events {
		bundle, ok := bundles[event.AssociatedBundle]
		if !ok {
			return nil, nil, fmt.Errorf("event resource with id: %s has unknown associated bundle with id: %s", event.In.ID, event.AssociatedBundle)
		}
		bundle.In.EventDefinitions = append(bundle.In.EventDefinitions, event.In)
	}

	resultBundles := make([]*BundleInputWithAssociatedPackages, 0)
	for _, bundle := range bundles {
		resultBundles = append(resultBundles, bundle)
	}

	return pkgs, resultBundles, nil
}

func rewriteRelativeURLsInJson(j json.RawMessage, baseURL, jsonPath string) (json.RawMessage, error) {
	parsedJson := gjson.ParseBytes(j)
	if parsedJson.IsArray() {
		items := make([]interface{}, 0, 0)
		for _, jsonElement := range parsedJson.Array() {
			rewrittenElement, err := rewriteRelativeURLsInJson(json.RawMessage(jsonElement.Raw), baseURL, jsonPath)
			if err != nil {
				return nil, err
			}
			m := make(map[string]interface{})
			if err := json.Unmarshal(rewrittenElement, &m); err != nil {
				return nil, err
			}
			items = append(items, m)
		}
		return json.Marshal(items)
	} else if parsedJson.IsObject() {
		urlProperty := gjson.GetBytes(j, jsonPath)
		if urlProperty.Exists() && !IsUrl(urlProperty.String()) {
			if len(baseURL) == 0 {
				return nil, fmt.Errorf("%s should be absolute URL if baseURL not provided in the document", jsonPath)
			} else {
				return sjson.SetBytes(j, jsonPath, baseURL+urlProperty.String())
			}
		}
	}
	return j, nil
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

func normalizeURL(url string) (string, error) {
	if len(url) > 0 && !IsUrl(url) {
		return "", fmt.Errorf("url %s is not a valid url", url)
	}
	return strings.TrimSuffix(url, "/"), nil
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
