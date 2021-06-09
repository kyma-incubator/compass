## Design e2e tests execution on real environment

### Installation of external services mock

* To install external services mock - compass installer can be used with some overrides like:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    installer: overrides
    component: compass
  annotations:
    strategy.spinnaker.io/replace: "false"
  name: compass-overrides-e2e-tests
  namespace: compass-installer
data:
  global.externalServicesMock.enabled: "true"
  global.externalServicesMock.auditlog: "false"
```

* Then starting only the installation will install external services mock
* After tests are executed the configmap with overrides can be deleted
* Run installer again which will remove the external services mock
* `global.externalServicesMock.auditlog: "false"` this is needed because we want to disable the auditlog configurations around the external services mock as it will not be tested in real environment
* some resources like secrets from external services mock chart should be guarded and not created in case the auditlog is disabled when installing external services mock


### Admin test user

* When e2e tests are started they should create the test tenants as they are not created by default in real environments. Perhaps use the tenants that are in the values.yaml. Possibly new ones can be created on the fly, but tests should be adapted to use the newly created tenants
* Dex static users can be added and used for test purposes. These users (one user?) can be associated with only test tenants, so it doesn't have any permissions on other tenants. However the addition of static users will slightly change the login UI and present the user the option to choose between login with static user or with the currently configured dex connector.

### Test tenants
* Test tenants can be loaded by using the tenant loader.
* The same tenants should be configured in director with the corresponding test admin user.
