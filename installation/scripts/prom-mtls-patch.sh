function promMtlsPatch() {
  patchPrometheusForMTLS
  patchAlertManagerForMTLS
  patchKymaServiceMonitorsForMTLS
  removeKymaPeerAuthsForPrometheus
  patchMonitoringTests
}

function patchPrometheusForMTLS() {
  patch=`cat <<"EOF"
apiVersion: monitoring.coreos.com/v1
kind: Prometheus
metadata:
  name: monitoring-prometheus
  namespace: kyma-system
spec:
  alerting:
    alertmanagers:
      - apiVersion: v2
        name: monitoring-alertmanager
        namespace: kyma-system
        pathPrefix: /
        port: web
        scheme: https
        tlsConfig:
          caFile: /etc/prometheus/secrets/istio.default/root-cert.pem
          certFile: /etc/prometheus/secrets/istio.default/cert-chain.pem
          keyFile: /etc/prometheus/secrets/istio.default/key.pem
          insecureSkipVerify: true
  podMetadata:
    annotations:
      sidecar.istio.io/inject: "true"
      traffic.sidecar.istio.io/includeInboundPorts: ""   # do not intercept any inbound ports
      traffic.sidecar.istio.io/includeOutboundIPRanges: ""  # do not intercept any outbound traffic
      proxy.istio.io/config: |
        # configure an env variable OUTPUT_CERTS to write certificates to the given folder
        proxyMetadata:
          OUTPUT_CERTS: /etc/istio-output-certs
      sidecar.istio.io/userVolumeMount: '[{"name": "istio-certs", "mountPath": "/etc/istio-output-certs"}]' # mount the shared volume at sidecar proxy
  volumes:
    - emptyDir:
        medium: Memory
      name: istio-certs
  volumeMounts:
    - mountPath: /etc/prometheus/secrets/istio.default/
      name: istio-certs
EOF
  `

  echo "${patch}" > patch.yaml
  kubectl apply -f patch.yaml
  rm patch.yaml
}

function patchAlertManagerForMTLS() {
  patch=`cat <<"EOF"
apiVersion: monitoring.coreos.com/v1
kind: Alertmanager
metadata:
  name: monitoring-alertmanager
  namespace: kyma-system
spec:
  podMetadata:
    annotations:
      sidecar.istio.io/inject: "true"
EOF
  `

  echo "${patch}" > patch.yaml
  kubectl apply -f patch.yaml
  rm patch.yaml
}

function patchKymaServiceMonitorsForMTLS() {
  kymaSvcMonitors=(kiali logging-fluent-bit logging-loki ory-oathkeeper-maester ory-hydra-maester tracing-jaeger-operator tracing-jaeger monitoring-grafana monitoring-alertmanager)

  crd="servicemonitors.monitoring.coreos.com"
  namespace="kyma-system"
  patchContent=`cat <<"EOF"
  - scheme: https
    tlsConfig:
      caFile: /etc/prometheus/secrets/istio.default/root-cert.pem
      certFile: /etc/prometheus/secrets/istio.default/cert-chain.pem
      keyFile: /etc/prometheus/secrets/istio.default/key.pem
      insecureSkipVerify: true
EOF
  `

  echo "$patchContent" > tmp_patch_content.yaml

  for sm in ${kymaSvcMonitors[@]}; do
    kubectl get ${crd} -n ${namespace} ${sm} -o yaml > ${sm}.yaml

    if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' -e '/ endpoints:/r tmp_patch_content.yaml' ${sm}.yaml
      sed -i '' -e 's/- port:/  port:/g' ${sm}.yaml
      sed -i '' -e 's/- metricRelabelings:/  metricRelabelings:/g' ${sm}.yaml
    else # assume Linux otherwise
      sed -i '/ endpoints:/r tmp_patch_content.yaml' ${sm}.yaml
      sed -i 's/- port:/  port:/g' ${sm}.yaml
      sed -i 's/- metricRelabelings:/  metricRelabelings:/g' ${sm}.yaml
    fi

    kubectl apply -f ${sm}.yaml || true

    rm ${sm}.yaml
  done

  rm tmp_patch_content.yaml
}

function removeKymaPeerAuthsForPrometheus() {
  crd="peerauthentications.security.istio.io"
  namespace="kyma-system"

  allPAs=(kiali logging-fluent-bit-metrics logging-loki monitoring-grafana-policy ory-oathkeeper-maester-metrics ory-hydra-maester-metrics tracing-jaeger-operator-metrics tracing-jaeger-metrics)

  for pa in ${allPAs[@]}; do
    kubectl delete ${crd} -n ${namespace} ${pa} || true
  done
}

function patchMonitoringTests() {
  crd="testdefinitions"
  namespace="kyma-system"
  name="monitoring"

  patchSidecarContainerCommand=`cat <<"EOF"
        - until curl -fsI http://localhost:15021/healthz/ready; do echo \"Waiting
          for Sidecar...\"; sleep 3; done; echo \"Sidecar available. Running the command...\";
          ./test-monitoring; x=$(echo $?); curl -fsI -X POST http://localhost:15020/quitquitquit
          && exit $x
EOF
  `

  echo "${patchSidecarContainerCommand}" > patchSidecarContainerCommand.yaml
  kubectl get ${crd} -n ${namespace} ${name} -o yaml > testdef.yaml

  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' -e 's/sidecar.istio.io\/inject: "false"/sidecar.istio.io\/inject: "true"/g' testdef.yaml
    sed -i '' -e '/- .\/test-monitoring/r patchSidecarContainerCommand.yaml' testdef.yaml
    sed -i '' -e 's/- .\/test-monitoring//g' testdef.yaml
  else # assume Linux otherwise
    sed -i 's/sidecar.istio.io\/inject: "false"/sidecar.istio.io\/inject: "true"/g' testdef.yaml
    sed -i '/- .\/test-monitoring/r patchSidecarContainerCommand.yaml' testdef.yaml
    sed -i 's/- .\/test-monitoring//g' testdef.yaml
  fi

  kubectl apply -f testdef.yaml || true

  rm testdef.yaml
  rm patchSidecarContainerCommand.yaml
}
