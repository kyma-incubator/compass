## Overview

To contribute to this project, follow the rules from the general [CONTRIBUTING.md](https://github.com/kyma-project/community/blob/master/CONTRIBUTING.md) document in the `community` repository, located in the `kyma-project` organization.

## Pull request compass rules
If you open pull request in compass repository and all test pass, 
you have to open draft pull request with updated compass components images to check if changes are compatible with Kyma.
In [values.yaml](https://github.com/kyma-project/kyma/blob/master/resources/compass/values.yaml) 
find components which are changed by your pull request and then update `version` to `PR-${Pull Request Number}` (e.g. `PR-123`).
Remember to give the link in compass pull request description to kyma pull request.
