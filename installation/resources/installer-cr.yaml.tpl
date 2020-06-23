apiVersion: "installer.kyma-project.io/v1alpha1"
kind: Installation
metadata:
  name: compass-installation
  namespace: default
  labels:
    action: install
    kyma-project.io/installation: ""
  finalizers:
    - finalizer.installer.kyma-project.io
spec:
  version: "__VERSION__"
  url: "__URL__"
  components:
    - name: "compass"
      namespace: "compass-system"
    - name: "istio-gateway"
      namespace: "istio-system"
    - name: "kyma-environment-broker"
      namespace: "kcp-system"
    - name: "oidc-kubeconfig-service"
      namespace: "kyma-system"
