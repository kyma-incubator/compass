package puller

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/open_discovery"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"net/http"
	"regexp"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) FetchOpenDiscoveryDocuments(url string) (open_discovery.Documents, error) {
	resp, err := http.Get(url + "/.well-known/open-discovery")
	if err != nil {
		return nil, errors.Wrap(err, "error fetching open discovery configuration")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body")
	}

	config := open_discovery.WellKnownConfig{}
	if err := json.Unmarshal(bodyBytes, &config); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling json body")
	}

	docs := make(open_discovery.Documents, 0, 0)
	for _, config := range config.OpenDiscoveryV1Config.DocumentConfigs {
		doc, err := c.FetchOpenDiscoveryDocument(url + config.URL)
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching OD document %s", url + config.URL)
		}
		docs = append(docs, *doc)
	}

	return docs, nil
}

func (c *Client) FetchOpenDiscoveryDocument(documentURL string) (*open_discovery.Document, error) {
	resp, err := http.Get(documentURL)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching open discovery document")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading document body")
	}
	documentJSON, err := c.parseExtensions(bodyBytes)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing extensions")
	}
	result := &open_discovery.Document{}
	if err := json.Unmarshal(documentJSON, &result); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling document")
	}
	return result, nil
}

func (c *Client) parseExtensions(body []byte) ([]byte, error) {
	extensionsRegex, err := regexp.Compile("^x-*") // TODO: Fix example to comply with the real regex: ^x-[\w\d.-_]+$
	if err != nil {
		return nil, err
	}
	jsonMap := make(map[string]interface{})
	if err := json.Unmarshal(body, &jsonMap); err != nil {
		return nil, err
	}
	extensions := make(map[string]interface{})
	for key, value := range jsonMap {
		if extensionsRegex.MatchString(key) {
			extensions[key] = value
		}
	}

	body, err = sjson.SetBytes(body, "extensions", extensions)
	if err != nil {
		return nil, err
	}

	for key, _ := range extensions {
		body, err = sjson.DeleteBytes(body, key)
		if err != nil {
			return nil, err
		}
	}

	packagesJSON := gjson.GetBytes(body, "packages")
	if packagesJSON.Exists() && packagesJSON.IsArray() {
		items := make([]interface{}, 0, 0)
		for _, item := range packagesJSON.Array() {
			parsedItem, err := c.parseExtensions([]byte(item.Raw))
			if err != nil {
				return nil, errors.Wrap(err, "error parsing package extensions")
			}
			m := make(map[string]interface{})
			if err := json.Unmarshal(parsedItem, &m); err != nil {
				return nil, err
			}
			items = append(items, m)
		}
		body, err = sjson.SetBytes(body, "packages", items)
		if err != nil {
			return nil, err
		}
	}

	bundlesJSON := gjson.GetBytes(body, "bundles")
	if bundlesJSON.Exists() && bundlesJSON.IsArray() {
		items := make([]interface{}, 0, 0)
		for _, item := range bundlesJSON.Array() {
			parsedItem, err := c.parseExtensions([]byte(item.String()))
			if err != nil {
				return nil, errors.Wrap(err, "error parsing package extensions")
			}
			m := make(map[string]interface{})
			if err := json.Unmarshal(parsedItem, &m); err != nil {
				return nil, err
			}
			items = append(items, m)
		}
		body, err = sjson.SetBytes(body, "bundles", items)
		if err != nil {
			return nil, err
		}
	}

	apiResourcesJSON := gjson.GetBytes(body, "apiResources")
	if apiResourcesJSON.Exists() && apiResourcesJSON.IsArray() {
		items := make([]interface{}, 0, 0)
		for _, item := range apiResourcesJSON.Array() {
			parsedItem, err := c.parseExtensions([]byte(item.String()))
			if err != nil {
				return nil, errors.Wrap(err, "error parsing package extensions")
			}
			m := make(map[string]interface{})
			if err := json.Unmarshal(parsedItem, &m); err != nil {
				return nil, err
			}
			items = append(items, m)
		}
		body, err = sjson.SetBytes(body, "apiResources", items)
		if err != nil {
			return nil, err
		}
	}

	eventResourcesJSON := gjson.GetBytes(body, "eventResources")
	if eventResourcesJSON.Exists() && eventResourcesJSON.IsArray() {
		items := make([]interface{}, 0, 0)
		for _, item := range eventResourcesJSON.Array() {
			parsedItem, err := c.parseExtensions([]byte(item.String()))
			if err != nil {
				return nil, errors.Wrap(err, "error parsing package extensions")
			}
			m := make(map[string]interface{})
			if err := json.Unmarshal(parsedItem, &m); err != nil {
				return nil, err
			}
			items = append(items, m)
		}
		body, err = sjson.SetBytes(body, "eventResources", items)
		if err != nil {
			return nil, err
		}
	}

	return body, nil
}