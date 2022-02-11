apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-installation-gke-benchmark-overrides
  namespace: compass-installer
  labels:
    component: compass
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.externalServicesMock.enabled: "true"
  gateway.gateway.auditlog.enabled: "false"
  gateway.gateway.auditlog.authMode: "oauth"
  global.systemFetcher.enabled: "true"
  global.systemFetcher.systemsAPIEndpoint: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/systemfetcher/systems"
  global.systemFetcher.systemsAPIFilterCriteria: "no"
  global.systemFetcher.systemsAPIFilterTenantCriteriaPattern: "tenant=%s"
  global.systemFetcher.systemToTemplateMappings: '[{"Name": "temp1", "SourceKey": ["prop"], "SourceValue": ["val1"] },{"Name": "temp2", "SourceKey": ["prop"], "SourceValue": ["val2"] }]'
  global.kubernetes.serviceAccountTokenJWKS: "https://container.googleapis.com/v1beta1/projects/$CLOUDSDK_CORE_PROJECT/locations/$CLOUDSDK_COMPUTE_ZONE/clusters/$COMMON_NAME/jwks"
  global.oathkeeper.mutators.authenticationMappingServices.tenant-fetcher.authenticator.enabled: "true"
  global.oathkeeper.mutators.authenticationMappingServices.subscriber.authenticator.enabled: "true"
  system-broker.http.client.skipSSLValidation: "true"
  operations-controller.http.client.skipSSLValidation: "true"
  global.systemFetcher.http.client.skipSSLValidation: "true"
  global.http.client.skipSSLValidation: "true"
  global.tests.http.client.skipSSLValidation.director: "true"
  global.tests.http.client.skipSSLValidation.ordService: "true"
