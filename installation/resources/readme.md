# Compass Resources

## Installation CRs

- [Kyma Dependencies](installer-cr-kyma-dependencies.yaml)
Kyma components that are required for Compass Installation

- [Compass installation](.installer-cr-yaml.tpl)
Components installed by the Compas Installer

## KYMA_VERSION file

KYMA_VERSION is a file used to configure the version of Kyma used during the Compass installation.

### Possible values

| Value                   	| Example Value     	| Explanation             	|
|-------------------------	|-------------------	|-------------------------	|
| `master`                	| `master`          	| Latest master version   	|
| `master-${commit hash}` 	| `master-34edf09a` 	| Specific master version 	|
| `PR-${PR number}`       	| `PR-1420`         	| Specific PR version     	|
| `${Release number}`     	| `1.13.0`          	| Release version         	|
