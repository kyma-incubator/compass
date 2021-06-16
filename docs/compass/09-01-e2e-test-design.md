## Design e2e tests execution on real environment

### Installation of external services mock

1. To install external services mock, you can use the Compass installer with some overrides like the following:

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
Since we want to disable the auditlog configurations around the external services mock, as it will not be tested in real environment, it is requred to set  `global.externalServicesMock.auditlog: "false"`.  
When, you start the Compass installer again on an already existing Compass installation, it only installs the external services mock as an addition to the existing installation.

2. After the tests are carried out, you must delete the `ConfigMap` resource with the overrides.
3. To remove the external services mock, run the Compass installer again. 


Note that some resources, such as, secrets from external services mock chart, must be guarded and not created in case the auditlog is disabled when installing external services mock.

### Test tenants
1. You can load test tenants by using the tenant loader.
2. The same tenants must be configured in director with the corresponding test admin user.
