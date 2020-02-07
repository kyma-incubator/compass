package mock

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func FixPackage() *graphql.PackageDefinition {
	desc := "Lorem ipsum"
	jsonSchema := `{
  "definitions": {}, 
  "$schema": "http://json-schema.org/draft-07/schema#", 
  "$id": "http://example.com/root.json", 
  "type": "object", 
  "title": "The Root Schema", 
  "description": "An explanation about the purpose of this instance.", 
  "readOnly": true, 
  "writeOnly": false, 
  "required": [
    "checked", 
    "dimensions", 
    "id", 
    "name", 
    "price", 
    "tags"
  ], 
  "properties": {
    "checked": {
      "$id": "#/properties/checked", 
      "type": "boolean", 
      "title": "The Checked Schema", 
      "description": "An explanation about the purpose of this instance.", 
      "default": false, 
      "examples": [
        false
      ], 
      "readOnly": true, 
      "writeOnly": false
    }, 
    "dimensions": {
      "$id": "#/properties/dimensions", 
      "type": "object", 
      "title": "The Dimensions Schema", 
      "description": "An explanation about the purpose of this instance.", 
      "readOnly": true, 
      "writeOnly": false, 
      "required": [
        "width", 
        "height"
      ], 
      "properties": {
        "width": {
          "$id": "#/properties/dimensions/properties/width", 
          "type": "integer", 
          "title": "The Width Schema", 
          "description": "An explanation about the purpose of this instance.", 
          "default": 0, 
          "examples": [
            5
          ], 
          "readOnly": true, 
          "writeOnly": false
        }, 
        "height": {
          "$id": "#/properties/dimensions/properties/height", 
          "type": "integer", 
          "title": "The Height Schema", 
          "description": "An explanation about the purpose of this instance.", 
          "default": 0, 
          "examples": [
            10
          ], 
          "readOnly": true, 
          "writeOnly": false
        }
      }
    }, 
    "id": {
      "$id": "#/properties/id", 
      "type": "integer", 
      "title": "The Id Schema", 
      "description": "An explanation about the purpose of this instance.", 
      "default": 0, 
      "examples": [
        1
      ], 
      "readOnly": true, 
      "writeOnly": false
    }, 
    "name": {
      "$id": "#/properties/name", 
      "type": "string", 
      "title": "The Name Schema", 
      "description": "An explanation about the purpose of this instance.", 
      "default": "", 
      "examples": [
        "A green door"
      ], 
      "readOnly": true, 
      "writeOnly": false, 
      "pattern": "^(.*)$"
    }, 
    "price": {
      "$id": "#/properties/price", 
      "type": "number", 
      "title": "The Price Schema", 
      "description": "An explanation about the purpose of this instance.", 
      "default": 0.0, 
      "examples": [
        12.5
      ], 
      "readOnly": true, 
      "writeOnly": false
    }, 
    "tags": {
      "$id": "#/properties/tags", 
      "type": "array", 
      "title": "The Tags Schema", 
      "description": "An explanation about the purpose of this instance.", 
      "readOnly": true, 
      "writeOnly": false, 
      "items": {
        "$id": "#/properties/tags/items", 
        "type": "string", 
        "title": "The Items Schema", 
        "description": "An explanation about the purpose of this instance.", 
        "default": "", 
        "examples": [
          "home", 
          "green"
        ], 
        "readOnly": true, 	
        "writeOnly": false, 
        "pattern": "^(.*)$"
      }
    }
  }
}`
	gqlJSONSchema := graphql.JSONSchema(jsonSchema)

	return &graphql.PackageDefinition{
		ID:                    "a6a64619-aa28-4d5d-bbe8-86ab4b817abf",
		Name:                  "Lorem ipsum dolores sit amet",
		Description:           &desc,
		AuthRequestJSONSchema: &gqlJSONSchema,
		DefaultAuth: &graphql.Auth{
			Credential: graphql.BasicCredentialData{
				Username: "foo",
				Password: "bar",
			},
		},
	}
}
