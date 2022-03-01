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
  global.kubernetes.serviceAccountTokenIssuer: "https://container.googleapis.com/v1beta1/projects/$CLOUDSDK_CORE_PROJECT/locations/$CLOUDSDK_COMPUTE_ZONE/clusters/$COMMON_NAME"
  global.kubernetes.serviceAccountTokenJWKS: "https://container.googleapis.com/v1beta1/projects/$CLOUDSDK_CORE_PROJECT/locations/$CLOUDSDK_COMPUTE_ZONE/clusters/$COMMON_NAME/jwks"
  global.oathkeeper.mutators.authenticationMappingServices.tenant-fetcher.authenticator.enabled: "true"
  global.oathkeeper.mutators.authenticationMappingServices.subscriber.authenticator.enabled: "true"
  system-broker.http.client.skipSSLValidation: "true"
  operations-controller.http.client.skipSSLValidation: "true"
  global.systemFetcher.http.client.skipSSLValidation: "true"
  global.http.client.skipSSLValidation: "true"
  global.tests.http.client.skipSSLValidation.director: "true"
  global.tests.http.client.skipSSLValidation.ordService: "true"
  global.connector.caKey: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBc2IxNWNYWENVdkNGOGNDdjMzYi94ZCtzbW5tUjRaLzZmZlhXMEtGbHBPNU5rbTJaCjZPNElLS1pvNXdxL3dHMjE4dVZYbGVQNWFlL2s0V2RNakcwN2hIY3NHczQrSWZaV0g3L0QwWFBHc0pIWDFsejMKaWNHN3U0cXFpV09USTdwK3RMMTM5dUxiUzhST3BWWEs3bHdrK0ZXRVE5OW9sNzErQThVMkJxVDhXdmhsVEdGYQpWZmNWSUg4TEpSdDFSZHYyeWRkUDl6R2dMSmNFblkvRisyYndpY1JZbVBIOTZiU0NPcVBsU3cvZDQrVEcxekJXClJEUFBNQiszMW1EMVFUTzQ3UlFUdEt3ZVphVXRualZoa3BURVRGOUd1M0tiWjZkT2tJRXFJWjFTVVpXdVJWblIKcUJ0R3JseXdXNVRONytGMkNJbkJOVGxVcVBrc3A4NUZVVVNxVFFJREFRQUJBb0lCQUVoV3RrT2dTdHVZdXRzZQpzalcvNSs5dnpuNzhkWXdmb1VKOHVOWW1xZ2pMV0ZUOU9JUGR4UUpPWUNtUWJXUnpBbmQrTWZ4MlVYOTFQSEVrCnFyb0lodzJ0dHd5ZDNoblNlVkRvcWxqbnh2ajhFcDFUTHdncENqQVZDcjFxQW11ckxvQi9FSUV4NlZEWDc2NUkKMFpQYmVzeDdlWjVxSWRhSUwrNTI2RHNpRVBjd3JjQXZjYnhaYmFCZmRUdVpoWHlpUjJRK3ErYkJNKzhla2NrNQpJc0ZHWVJMRk5oTDZXSHM3Q1Fzck5JMVQ1UERCR1hDUWsrMC9laWFIdnRPeWZaS2NNVy94Q0VpUkEzbWZzRHJGCm95WnFrMXFaTXZGdU1iRHNKajdDN2dSN2w3UlJHbm9BTGdmdmNHMlJSY1ArQWsrMU5BL096K0hCZ1RrSjh6WG8KMGJCTTZRRUNnWUVBM1haV1dyODEzT0RtYzdPemFtT096WUNSSVBkNnRUMTkrYkVtanByVnI3OVlkZUp5OExXTgpNSGU2Tlp3REtLWk1FTURaUHNINi9Jczl2TCtoa3pSd0NYRUhWa2dVczVYSWNvSERiWktVTXlZZGQ4cC9ESi8yCndybFNKRi9XOHV3YzVGcUY3SHdvZm1kYSt5ekhVSExDK2hwblFrR0g5dGtTV1VuTVdRMW9mTUVDZ1lFQXpYV1QKQy8yNkpYeWVHRWhBOWJOSTBESDE3NFpxNHpiZFgrSnVISUplaVE5alRvV3lqR3p5ckZvd2tXK3VZM3NqQU1MQQo1SE5TaW5nWnpEZ01jenFUbWJjMG9yQlVUMEF6dFBSZHg0cjRZZUVGL1h2dlJxYnNyeGdHTWliclhyd3BDSjVDCnFZVk82RWtqNXJiUWJKY0NQL1EyR1BFZ1pLR0JMdk0wSzd2UTlJMENnWUJ0aWpyc1orZWNlU0dEMlQ3RlFMbEIKckhZY2VFeVptUERXc0drQjRGUVJ1Zk5uVzdxK2xRNWhDdGR2N05zaklCNC9xeVBKaHVrK1FTRW9XeUR3VHQrYgp5K3gxSVBJY1lkbmp5WXVBaHlBR3JMT21yT0pxdkRTeDNEaGxCWUtzWlYxbEZlRm9ONEZRQkk5Yjdhb29nSnN3CldoNzVCckRaeUVUckpUV09Wck40QVFLQmdBeDZNMi9xL0w4Q0RtZlRHMzdRWUgra1NSYyt4b2I5OGZ1OHVJc3EKcjZzTE1EQzRsZHRKVW9OOUJxNE9aanpNWVpmT1BBQ2pzRU9RZjZDVFZzNDRwSFlWVmpEN0hHT2p0b0FxeHZjegpUVnBFWENURXZnZEFZK2RPUWpJUmd3SEIwNHdlY0ZYekxwT1V2WVZwWE1iN2RMdUZqVE4ra1VVTE9ka3NFK01FCkNQQ1JBb0dBQXEzUUptNUhzVzV1VEhSanlzRGM3UmdWZ2RhUTdRdTJHYVVkZEU0ZnJSMXB6SE1JRHhzclpaeWwKZ2lvcUNwUDd3VFV0Wmo0ZXQ3QmpGWU5nUUkvSEp6L3htNmJXYXRhSFBCR3grc2tQN3NLaG5ab3lsbzdDeThCLwp1cWRtT2FIaW9YSGdQc2lCcks0dlZKZnJ4aTlSbGhkWjJyeFpKNWJxYWdIcmdzdE1LUzg9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
  global.connector.caCertificate: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNwakNDQVk0Q0NRRDRIRW9Od21MSjhqQU5CZ2txaGtpRzl3MEJBUXNGQURBVU1SSXdFQVlEVlFRRERBbHMKYjJOaGJHaHZjM1F3SUJjTk1qSXdNakEzTVRVd056VTRXaGdQTWpFeU1qQXhNVFF4TlRBM05UaGFNQlF4RWpBUQpCZ05WQkFNTUNXeHZZMkZzYUc5emREQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFMRzllWEYxd2xMd2hmSEFyOTkyLzhYZnJKcDVrZUdmK24zMTF0Q2haYVR1VFpKdG1lanVDQ2ltYU9jS3Y4QnQKdGZMbFY1WGorV252NU9GblRJeHRPNFIzTEJyT1BpSDJWaCsvdzlGenhyQ1IxOVpjOTRuQnU3dUtxb2xqa3lPNgpmclM5ZC9iaTIwdkVUcVZWeXU1Y0pQaFZoRVBmYUplOWZnUEZOZ2FrL0ZyNFpVeGhXbFgzRlNCL0N5VWJkVVhiCjlzblhUL2N4b0N5WEJKMlB4ZnRtOEluRVdKangvZW0wZ2pxajVVc1AzZVBreHRjd1ZrUXp6ekFmdDlaZzlVRXoKdU8wVUU3U3NIbVdsTFo0MVlaS1V4RXhmUnJ0eW0yZW5UcENCS2lHZFVsR1Zya1ZaMGFnYlJxNWNzRnVVemUvaApkZ2lKd1RVNVZLajVMS2ZPUlZGRXFrMENBd0VBQVRBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQVFFQVdJb2pnUEhVCnZ1YkF0RWF3ajErY1Blcm8wdkVaRlJvNmFVK05idjJ2Y2J3M1RmczY0QlJabTQwOE05cysyRzNDSFdaRE9TdlgKY0hWdXo4eC9XSXZVZU53WmJNRVF4TVZHVVl1K1FrVnU4bnI1Z1daRkIranphbHhHVUxyNXRnbWpNYVQ2Zk9LcQp5MzNMUXFXZDhadFA2R3ZJS01SblhnelBtaEJ5MVhvVDZJZXVZVzFhRWI5Q3ZCNlRjOGlUK0lDaWlKL0pPWjFnClhOU1pWaUxnbGlkQnJlUm5ZZGtvalZxZzdDaTlKSGFIY2hITDBnS0xxeUdodFAycTlBMDVscXVzd1lNNm14dUwKM1dHaFRUK3kxSWRQMnQyUmRQc2xiRURvaGhMTHpTYkdEckR2VnMxc05NbjJJbmpmS2cvaVdCUlkrL3FCdU4xcwpBK2xXZkhDZjh4d2pUZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
  global.externalCertConfiguration.secrets.externalCertSvcSecret.manage: "true"
  global.externalServicesMock.oauthSecret.manage: "true"
  global.tests.basicCredentials.manage: "true"
  global.tests.ordService.subscriptionOauthSecret.manage: "true"

  global.connector.caKey: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBc2IxNWNYWENVdkNGOGNDdjMzYi94ZCtzbW5tUjRaLzZmZlhXMEtGbHBPNU5rbTJaCjZPNElLS1pvNXdxL3dHMjE4dVZYbGVQNWFlL2s0V2RNakcwN2hIY3NHczQrSWZaV0g3L0QwWFBHc0pIWDFsejMKaWNHN3U0cXFpV09USTdwK3RMMTM5dUxiUzhST3BWWEs3bHdrK0ZXRVE5OW9sNzErQThVMkJxVDhXdmhsVEdGYQpWZmNWSUg4TEpSdDFSZHYyeWRkUDl6R2dMSmNFblkvRisyYndpY1JZbVBIOTZiU0NPcVBsU3cvZDQrVEcxekJXClJEUFBNQiszMW1EMVFUTzQ3UlFUdEt3ZVphVXRualZoa3BURVRGOUd1M0tiWjZkT2tJRXFJWjFTVVpXdVJWblIKcUJ0R3JseXdXNVRONytGMkNJbkJOVGxVcVBrc3A4NUZVVVNxVFFJREFRQUJBb0lCQUVoV3RrT2dTdHVZdXRzZQpzalcvNSs5dnpuNzhkWXdmb1VKOHVOWW1xZ2pMV0ZUOU9JUGR4UUpPWUNtUWJXUnpBbmQrTWZ4MlVYOTFQSEVrCnFyb0lodzJ0dHd5ZDNoblNlVkRvcWxqbnh2ajhFcDFUTHdncENqQVZDcjFxQW11ckxvQi9FSUV4NlZEWDc2NUkKMFpQYmVzeDdlWjVxSWRhSUwrNTI2RHNpRVBjd3JjQXZjYnhaYmFCZmRUdVpoWHlpUjJRK3ErYkJNKzhla2NrNQpJc0ZHWVJMRk5oTDZXSHM3Q1Fzck5JMVQ1UERCR1hDUWsrMC9laWFIdnRPeWZaS2NNVy94Q0VpUkEzbWZzRHJGCm95WnFrMXFaTXZGdU1iRHNKajdDN2dSN2w3UlJHbm9BTGdmdmNHMlJSY1ArQWsrMU5BL096K0hCZ1RrSjh6WG8KMGJCTTZRRUNnWUVBM1haV1dyODEzT0RtYzdPemFtT096WUNSSVBkNnRUMTkrYkVtanByVnI3OVlkZUp5OExXTgpNSGU2Tlp3REtLWk1FTURaUHNINi9Jczl2TCtoa3pSd0NYRUhWa2dVczVYSWNvSERiWktVTXlZZGQ4cC9ESi8yCndybFNKRi9XOHV3YzVGcUY3SHdvZm1kYSt5ekhVSExDK2hwblFrR0g5dGtTV1VuTVdRMW9mTUVDZ1lFQXpYV1QKQy8yNkpYeWVHRWhBOWJOSTBESDE3NFpxNHpiZFgrSnVISUplaVE5alRvV3lqR3p5ckZvd2tXK3VZM3NqQU1MQQo1SE5TaW5nWnpEZ01jenFUbWJjMG9yQlVUMEF6dFBSZHg0cjRZZUVGL1h2dlJxYnNyeGdHTWliclhyd3BDSjVDCnFZVk82RWtqNXJiUWJKY0NQL1EyR1BFZ1pLR0JMdk0wSzd2UTlJMENnWUJ0aWpyc1orZWNlU0dEMlQ3RlFMbEIKckhZY2VFeVptUERXc0drQjRGUVJ1Zk5uVzdxK2xRNWhDdGR2N05zaklCNC9xeVBKaHVrK1FTRW9XeUR3VHQrYgp5K3gxSVBJY1lkbmp5WXVBaHlBR3JMT21yT0pxdkRTeDNEaGxCWUtzWlYxbEZlRm9ONEZRQkk5Yjdhb29nSnN3CldoNzVCckRaeUVUckpUV09Wck40QVFLQmdBeDZNMi9xL0w4Q0RtZlRHMzdRWUgra1NSYyt4b2I5OGZ1OHVJc3EKcjZzTE1EQzRsZHRKVW9OOUJxNE9aanpNWVpmT1BBQ2pzRU9RZjZDVFZzNDRwSFlWVmpEN0hHT2p0b0FxeHZjegpUVnBFWENURXZnZEFZK2RPUWpJUmd3SEIwNHdlY0ZYekxwT1V2WVZwWE1iN2RMdUZqVE4ra1VVTE9ka3NFK01FCkNQQ1JBb0dBQXEzUUptNUhzVzV1VEhSanlzRGM3UmdWZ2RhUTdRdTJHYVVkZEU0ZnJSMXB6SE1JRHhzclpaeWwKZ2lvcUNwUDd3VFV0Wmo0ZXQ3QmpGWU5nUUkvSEp6L3htNmJXYXRhSFBCR3grc2tQN3NLaG5ab3lsbzdDeThCLwp1cWRtT2FIaW9YSGdQc2lCcks0dlZKZnJ4aTlSbGhkWjJyeFpKNWJxYWdIcmdzdE1LUzg9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
  global.connector.caCertificate: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNwakNDQVk0Q0NRRDRIRW9Od21MSjhqQU5CZ2txaGtpRzl3MEJBUXNGQURBVU1SSXdFQVlEVlFRRERBbHMKYjJOaGJHaHZjM1F3SUJjTk1qSXdNakEzTVRVd056VTRXaGdQTWpFeU1qQXhNVFF4TlRBM05UaGFNQlF4RWpBUQpCZ05WQkFNTUNXeHZZMkZzYUc5emREQ0NBU0l3RFFZSktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCCkFMRzllWEYxd2xMd2hmSEFyOTkyLzhYZnJKcDVrZUdmK24zMTF0Q2haYVR1VFpKdG1lanVDQ2ltYU9jS3Y4QnQKdGZMbFY1WGorV252NU9GblRJeHRPNFIzTEJyT1BpSDJWaCsvdzlGenhyQ1IxOVpjOTRuQnU3dUtxb2xqa3lPNgpmclM5ZC9iaTIwdkVUcVZWeXU1Y0pQaFZoRVBmYUplOWZnUEZOZ2FrL0ZyNFpVeGhXbFgzRlNCL0N5VWJkVVhiCjlzblhUL2N4b0N5WEJKMlB4ZnRtOEluRVdKangvZW0wZ2pxajVVc1AzZVBreHRjd1ZrUXp6ekFmdDlaZzlVRXoKdU8wVUU3U3NIbVdsTFo0MVlaS1V4RXhmUnJ0eW0yZW5UcENCS2lHZFVsR1Zya1ZaMGFnYlJxNWNzRnVVemUvaApkZ2lKd1RVNVZLajVMS2ZPUlZGRXFrMENBd0VBQVRBTkJna3Foa2lHOXcwQkFRc0ZBQU9DQVFFQVdJb2pnUEhVCnZ1YkF0RWF3ajErY1Blcm8wdkVaRlJvNmFVK05idjJ2Y2J3M1RmczY0QlJabTQwOE05cysyRzNDSFdaRE9TdlgKY0hWdXo4eC9XSXZVZU53WmJNRVF4TVZHVVl1K1FrVnU4bnI1Z1daRkIranphbHhHVUxyNXRnbWpNYVQ2Zk9LcQp5MzNMUXFXZDhadFA2R3ZJS01SblhnelBtaEJ5MVhvVDZJZXVZVzFhRWI5Q3ZCNlRjOGlUK0lDaWlKL0pPWjFnClhOU1pWaUxnbGlkQnJlUm5ZZGtvalZxZzdDaTlKSGFIY2hITDBnS0xxeUdodFAycTlBMDVscXVzd1lNNm14dUwKM1dHaFRUK3kxSWRQMnQyUmRQc2xiRURvaGhMTHpTYkdEckR2VnMxc05NbjJJbmpmS2cvaVdCUlkrL3FCdU4xcwpBK2xXZkhDZjh4d2pUZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
