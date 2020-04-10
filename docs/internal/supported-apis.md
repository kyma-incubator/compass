# Supported API specifications

These are the currently supported API specification versions and formats in different components:

> **NOTE:** The last row describes support for extracting API specifications from archive files.

| | Console UI & Compass UI (rendering) | Runtime Rafter (storing) |  Compass UI (uploading) | Compass Backend (storing) | Runtime Agent (storing)
| --- | --- | --- | --- | --- | --- |
| OpenAPI<br>(YAML, JSON) | all formats, v2 and v3 | no validation / conversion | all versions and formats, validation by property | all versions and formats, no validation | all versions, JSON only, no validation and conversion |
| OData<br>(XML, JSON in v4) | XML only, all versions converted to v4 on client side |  no validation / conversion |  all versions, XML only, validation by parsing XML file and checking property | all versions and formats, no validation | all versions and formats, no validation and conversion |     
| AsyncAPI<br>YAML, JSON | all formats, v2 only | validation and conversion to JSON and v2 | all versions and formats, validation by property | all versions and formats, no validation | all versions, JSON only, no validation and conversion |
| Archives | no | yes (TAR and ZIP) | no | no | no |

## Planned support for API specification versions and formats

The following table shows the proposed supported API specification versions and formats in different components.
Rendering would support only a single version of API specifications. Rafter would be responsible for validating and converting API specifications for Console UI. Because Compass UI does not use Rafter, the conversion would have to be done in Compass Backend. Preferably Rafter's converters and validators should be exposed as services or libraries that could be used by the Compass Backend and Compass UI.

> **NOTE:**  The last row describes support for extracting API specifications from archive files.

| | Console UI & Compass UI (rendering) | Runtime Rafter (storing) |  Compass UI (uploading) | Compass Backend (storing) | Runtime Agent (storing)
| --- | --- | --- | --- | --- |  --- |
| OpenAPI<br>(YAML, JSON) | v3, JSON and YAML | validation and conversion to version supported by renderer | all versions and formats, validation using same logic as rafter | all versions and formats, validation and conversion using same logic as rafter | all versions and formats, no validation and conversion |
| OData<br>(XML, JSON in v4) | v4, JSON |  validation and conversion to version supported by renderer |  all versions and formats, validation using same logic as rafter | all versions and formats, validation and conversion using same logic as rafter | all versions and formats, no validation and conversion |
| AsyncAPI<br>YAML, JSON | v2, JSON and YAML | validation and conversion to version supported by renderer | all versions and formats, validation using same logic as rafter | all versions and formats, validation and conversion using same logic as rafter | all versions and formats, no validation and conversion |
| Archives | no | yes (TAR and ZIP) | no | no | no |
