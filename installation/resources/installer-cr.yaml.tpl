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
  # TODO: If ?
    - name: "cert-manager"
      namespace: "cert-manager"
  # END IF
    - name: "compass"
      namespace: "compass-system"