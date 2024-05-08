CURRENT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$CURRENT_DIR"/utils.sh

function prometheusMTLSPatch() {
  enableNodeExporterMTLS
  patchKymaServiceMonitorsForMTLS
  removeKymaPeerAuthsForPrometheus
}

function enableNodeExporterMTLS() {
  # The patches around the DaemonSet involve an addition of two init containers that together setup certificates
  # for the node-exporter application to use. There are also two new mounts - a shared directory (node-certs)
  # and the Istio CA secret (istio-certs).
  # This can be moved to the Helm values.yaml but it depends on the existence of Istio (its certificate Secret has to be
  # replicated in the kyma-system namespace as well). As Istio and the monitoring stack are both deployed by Kyma this
  # Secret replication is tricky, that's why the patch is kept.

  daemonset=$(cat <<"EOF"
spec:
  template:
    spec:
      initContainers:
      - name: certs-init
        image: emberstack/openssl:alpine-latest
        command: ['sh', '-c', 'openssl req -newkey rsa:2048 -nodes -days 365000 -subj "/CN=$(NODE_NAME)" -keyout /etc/certs/node.key -out /etc/certs/node.csr && openssl x509 -req -days 365000 -set_serial 01 -in /etc/certs/node.csr -out /etc/certs/node.crt -CA /etc/istio/certs/ca-cert.pem -CAkey /etc/istio/certs/ca-key.pem']
        env:
          - name: NODE_NAME
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
        volumeMounts:
        - name: istio-certs
          mountPath: /etc/istio/certs
          readOnly: true
        - name: node-certs
          mountPath: /etc/certs
      - name: web-config-init
        image: busybox:1.34.1
        command: ['sh', '-c', 'printf "tls_server_config:\\n  cert_file: /etc/certs/node.crt\\n  key_file: /etc/certs/node.key\\n  client_auth_type: \"RequireAndVerifyClientCert\"\\n  client_ca_file: /etc/istio/certs/ca-cert.pem" > /etc/certs/web.yaml']
        volumeMounts:
        - name: node-certs
          mountPath: /etc/certs
      containers:
      - args:
        - --web.config.file=/etc/certs/web.yaml
        name: node-exporter
        env:
        - name: HOST_IP
          value: 0.0.0.0
        livenessProbe: null
        readinessProbe: null
        volumeMounts:
        - mountPath: /etc/certs
          name: node-certs
        - name: istio-certs
          mountPath: /etc/istio/certs
      volumes:
      - name: istio-certs
        secret:
          secretName: istio-ca-secret
      - name: node-certs
        emptyDir: {}

EOF
  )
  echo "$daemonset" > daemonset.yaml

  kubectl_k3d_kyma get secret istio-ca-secret --namespace=istio-system -o yaml | grep -v '^\s*namespace:\s' | kubectl_k3d_kyma replace --force --namespace=kyma-system -f -

  kubectl_k3d_kyma patch daemonset monitoring-prometheus-node-exporter -n kyma-system --patch-file daemonset.yaml

  rm daemonset.yaml
} 

function patchKymaServiceMonitorsForMTLS() {
  # Some of the ServiceMonitor MTLS overrides were moved to the Kyma Helm chart overrides
  sm="monitoring-operator"
  crd="servicemonitors.monitoring.coreos.com"
  namespace="kyma-system"
  scheme="https"
  tlsConfig='{
    "caFile": "/etc/prometheus/secrets/istio.default/root-cert.pem",
    "certFile": "/etc/prometheus/secrets/istio.default/cert-chain.pem",
    "keyFile": "/etc/prometheus/secrets/istio.default/key.pem",
    "insecureSkipVerify": true
  }'

  if kubectl_k3d_kyma get ${crd} -n ${namespace} "${sm}" > /dev/null; then
    kubectl_k3d_kyma get ${crd} -n ${namespace} "${sm}" -o json > "${sm}.json"

    cp "${sm}.json" tmp.json
    jq --arg newSchema "$scheme" '.spec.endpoints[].scheme = $newSchema' tmp.json > "${sm}.json"
    cp "${sm}.json" tmp.json
    jq --argjson newTlsConfig "$tlsConfig" '.spec.endpoints[].tlsConfig = $newTlsConfig' tmp.json > "${sm}.json"
    rm tmp.json

    kubectl_k3d_kyma apply -f "${sm}.json" || true

    rm "${sm}.json"
  fi
}

function removeKymaPeerAuthsForPrometheus() {
  crd="peerauthentications.security.istio.io"

  kubectl_k3d_kyma delete ${crd} -n kyma-system monitoring-grafana-policy || true
}
