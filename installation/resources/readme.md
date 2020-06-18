#KYMA_VERSION file

KYMA_VERSION is a file used to configure the version of Kyma used during the Compass installation.

## Possible values

| Value                   	| Explanation             	|
|-------------------------	|-------------------------	|
| `master`                	| Latest master version   	|
| `master-${commit hash}` 	| Specific master version 	|
| `PR-${PR number}`       	| Specific PR version     	|
| Release eg. `1.14.0`    	| Release version         	|