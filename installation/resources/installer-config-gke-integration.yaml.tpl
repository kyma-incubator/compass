apiVersion: v1
kind: ConfigMap
metadata:
  name: compass-installation-gke-integration-overrides
  namespace: compass-installer
  labels:
    component: compass
    installer: overrides
    kyma-project.io/installation: ""
data:
  global.externalServicesMock.enabled: "true"
  global.externalServicesMock.auditlog.applyMockConfiguration: "true"
  gateway.gateway.auditlog.enabled: "true"
  gateway.gateway.auditlog.authMode: "oauth"
  global.systemFetcher.enabled: "true"
  global.systemFetcher.systemsAPIEndpoint: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/systemfetcher/systems"
  global.systemFetcher.systemsAPIFilterCriteria: "no"
  global.systemFetcher.systemsAPIFilterTenantCriteriaPattern: "tenant=%s"
  global.systemFetcher.systemToTemplateMappings: '[{"Name": "temp1", "SourceKey": ["prop"], "SourceValue": ["val1"] },{"Name": "temp2", "SourceKey": ["prop"], "SourceValue": ["val2"] }]'
  global.systemFetcher.oauth.skipSSLValidation: "true"
  global.migratorJob.nodeSelectorEnabled: "true"
  global.kubernetes.serviceAccountTokenJWKS: "https://container.googleapis.com/v1beta1/projects/$CLOUDSDK_CORE_PROJECT/locations/$CLOUDSDK_COMPUTE_ZONE/clusters/$COMMON_NAME/jwks"
  global.oathkeeper.mutators.authenticationMappingServices.tenant-fetcher.authenticator.enabled: "true"
  global.oathkeeper.mutators.authenticationMappingServices.subscriber.authenticator.enabled: "true"
  system-broker.http.client.skipSSLValidation: "true"
  connector.http.client.skipSSLValidation: "true"
  operations-controller.http.client.skipSSLValidation: "true"
  global.systemFetcher.http.client.skipSSLValidation: "true"
  global.ordAggregator.http.client.skipSSLValidation: "true"
  global.http.client.skipSSLValidation: "true"
  global.tests.http.client.skipSSLValidation.director: "true"
  global.tests.http.client.skipSSLValidation.ordService: "true"

  global.tenantFetchers.account-fetcher.enabled: "true"
  global.tenantFetchers.account-fetcher.dbPool.maxOpenConnections: "1"
  global.tenantFetchers.account-fetcher.dbPool.maxIdleConnections: "1"
  global.tenantFetchers.account-fetcher.manageSecrets: "true"
  global.tenantFetchers.account-fetcher.secret.name: "compass-account-fetcher-secret"
  global.tenantFetchers.account-fetcher.secret.clientIdKey: "client-id"
  global.tenantFetchers.account-fetcher.secret.clientSecretKey: "client-secret"
  global.tenantFetchers.account-fetcher.secret.oauthUrlKey: "url"
  global.tenantFetchers.account-fetcher.oauth.client: "client_id"
  global.tenantFetchers.account-fetcher.oauth.secret: "client_secret"
  global.tenantFetchers.account-fetcher.oauth.tokenURL: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/secured/oauth/token"
  global.tenantFetchers.account-fetcher.providerName: "external-svc-mock"
  global.tenantFetchers.account-fetcher.schedule: "0 23 * * *"
  global.tenantFetchers.account-fetcher.kubernetes.configMapNamespace: "compass-system"
  global.tenantFetchers.account-fetcher.kubernetes.pollInterval: "2s"
  global.tenantFetchers.account-fetcher.kubernetes.pollTimeout: "1m"
  global.tenantFetchers.account-fetcher.kubernetes.timeout: "2m"
  global.tenantFetchers.account-fetcher.fieldMapping.idField: "guid"
  global.tenantFetchers.account-fetcher.fieldMapping.nameField: "displayName"
  global.tenantFetchers.account-fetcher.fieldMapping.customerIdField: "customerId"
  global.tenantFetchers.account-fetcher.fieldMapping.discriminatorField: ""
  global.tenantFetchers.account-fetcher.fieldMapping.discriminatorValue: ""
  global.tenantFetchers.account-fetcher.fieldMapping.totalPagesField: "totalPages"
  global.tenantFetchers.account-fetcher.fieldMapping.totalResultsField: "totalResults"
  global.tenantFetchers.account-fetcher.fieldMapping.tenantEventsField: "events"
  global.tenantFetchers.account-fetcher.fieldMapping.detailsField: "eventData"
  global.tenantFetchers.account-fetcher.fieldMapping.entityTypeField: "type"
  global.tenantFetchers.account-fetcher.queryMapping.pageNumField: "page"
  global.tenantFetchers.account-fetcher.queryMapping.pageSizeField: "resultsPerPage"
  global.tenantFetchers.account-fetcher.queryMapping.timestampField: "ts"
  global.tenantFetchers.account-fetcher.query.startPage: "1"
  global.tenantFetchers.account-fetcher.query.pageSize: "1000"
  global.tenantFetchers.account-fetcher.shouldSyncSubaccounts: "false"
  global.tenantFetchers.account-fetcher.accountRegion: "local"
  global.tenantFetchers.account-fetcher.endpoints.accountCreated: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/tenant-fetcher/global-account-create"
  global.tenantFetchers.account-fetcher.endpoints.accountDeleted: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/tenant-fetcher/global-account-delete"
  global.tenantFetchers.account-fetcher.endpoints.accountUpdated: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/tenant-fetcher/global-account-update"

  global.tenantFetchers.subaccount-fetcher.enabled: "true"
  global.tenantFetchers.subaccount-fetcher.dbPool.maxOpenConnections: "1"
  global.tenantFetchers.subaccount-fetcher.dbPool.maxIdleConnections: "1"
  global.tenantFetchers.subaccount-fetcher.manageSecrets: "true"
  global.tenantFetchers.subaccount-fetcher.secret.name: "compass-subaccount-fetcher-secret"
  global.tenantFetchers.subaccount-fetcher.secret.clientIdKey: "client-id"
  global.tenantFetchers.subaccount-fetcher.secret.clientSecretKey: "client-secret"
  global.tenantFetchers.subaccount-fetcher.secret.oauthUrlKey: "url"
  global.tenantFetchers.subaccount-fetcher.oauth.client: "client_id"
  global.tenantFetchers.subaccount-fetcher.oauth.secret: "client_secret"
  global.tenantFetchers.subaccount-fetcher.oauth.tokenURL: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/secured/oauth/token"
  global.tenantFetchers.subaccount-fetcher.providerName: "subaccount-fetcher"
  global.tenantFetchers.subaccount-fetcher.schedule: "0 23 * * *"
  global.tenantFetchers.subaccount-fetcher.kubernetes.configMapNamespace: "compass-system"
  global.tenantFetchers.subaccount-fetcher.kubernetes.pollInterval: "2s"
  global.tenantFetchers.subaccount-fetcher.kubernetes.pollTimeout: "1m"
  global.tenantFetchers.subaccount-fetcher.kubernetes.timeout: "2m"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.idField: "guid"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.nameField: "displayName"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.customerIdField: "customerId"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.discriminatorField: ""
  global.tenantFetchers.subaccount-fetcher.fieldMapping.discriminatorValue: ""
  global.tenantFetchers.subaccount-fetcher.fieldMapping.totalPagesField: "totalPages"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.totalResultsField: "totalResults"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.tenantEventsField: "events"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.detailsField: "eventData"
  global.tenantFetchers.subaccount-fetcher.fieldMapping.entityTypeField: "type"
  global.tenantFetchers.subaccount-fetcher.queryMapping.pageNumField: "page"
  global.tenantFetchers.subaccount-fetcher.queryMapping.pageSizeField: "resultsPerPage"
  global.tenantFetchers.subaccount-fetcher.queryMapping.timestampField: "ts"
  global.tenantFetchers.subaccount-fetcher.query.startPage: "1"
  global.tenantFetchers.subaccount-fetcher.query.pageSize: "1000"
  global.tenantFetchers.subaccount-fetcher.shouldSyncSubaccounts: "true"
  global.tenantFetchers.subaccount-fetcher.accountRegion: "local"
  global.tenantFetchers.subaccount-fetcher.subaccountRegions: "test"
  global.tenantFetchers.subaccount-fetcher.endpoints.subaccountCreated: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/tenant-fetcher/subaccount-create"
  global.tenantFetchers.subaccount-fetcher.endpoints.subaccountDeleted: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/tenant-fetcher/subaccount-delete"
  global.tenantFetchers.subaccount-fetcher.endpoints.subaccountUpdated: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/tenant-fetcher/subaccount-update"
  global.tenantFetchers.subaccount-fetcher.endpoints.subaccountMoved: "http://compass-external-services-mock.compass-system.svc.cluster.local:8080/tenant-fetcher/subaccount-move"
