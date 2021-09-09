# Compass Resources

## Installation CRs

The following table lists all the Installation custom resource files for Compass:

| File                                	| Description                                                	|
|-------------------------------------	|------------------------------------------------------------	|
| `installer-cr-kyma-minimal.yaml` 	| Kyma components that are required for Compass installation. 	|
| `installer-cr.yaml.tpl`               	| Components installed by the Compass Installer.               	|

## KYMA_VERSION file

`KYMA_VERSION` is the file specifying the version of Kyma to use during the Compass installation.

### Possible values

| Value                   	| Example Value     	| Explanation             	|
|-------------------------	|-------------------	|-------------------------	|
| `main`                	| `main`          	| Latest main version.   	|
| `main-{COMMIT_HASH}` 	| `main-34edf09a` 	| Specific main version. 	|
| `PR-{PR_NUMBER}`       	| `PR-1420`         	| Specific PR version.     	|
| `{RELEASE_NUMBER}`     	| `1.13.0`          	| Release version.         	|
