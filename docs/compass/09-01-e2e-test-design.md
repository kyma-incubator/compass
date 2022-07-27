## Design e2e tests execution on real environment

### Installation of external services mock

1. To install external services mock, you can pass additional overrides file to the Compass installation script:

Save the following .yaml with external services mock overrides into a file (for example: externalServicesMockOverrides.yaml)
```yaml
global:
  externalServicesMock:
      enabled: true
      auditlog:
        applyMockConfiguration: false
```

And then, start the Compass installation by using the following command that installs external services mock too:

```bash
<script from ../../installation/scripts/install-compass.sh> --overrides-file <pass all override files used for initial Compass installation> --overrides-file <file from above step - e.g. externalServicesMockOverrides.yaml> --timeout <e.g: 30m0s>
```

Since we want to reuse the real auditlog configurations (configmaps & secrets) in the external services mock auditlog tests, it is requred to set  `global.externalServicesMock.auditlog.applyMockConfiguration: "false"`.  

2. To remove the external services mock, run the script for Compass installation and make sure that the external services mock overrides file is excluded:

```bash
<script from ../../installation/scripts/install-compass.sh> --overrides-file <pass all override files used for initial Compass installation> --timeout <e.g: 30m0s>
```

Note that some resources, such as, secrets from external services mock chart, must be guarded and not created in case the auditlog is disabled when installing external services mock.

### Test tenants
1. You can load test tenants by using the tenant loader.
2. The same tenants must be configured in director with the corresponding test admin user.
