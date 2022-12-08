function prometheusMTLSPatch() {
  patchPrometheusForMTLS
  patchAlertManagerForMTLS
  enableNodeExporterMTLS
  patchDeploymentsToInjectSidecar
  patchKymaServiceMonitorsForMTLS
  removeKymaPeerAuthsForPrometheus
}

function patchPrometheusForMTLS() {
  patch=$(cat <<"EOF"
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
        port: http-web
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
  )

  echo "${patch}" > patch.yaml
  kubectl apply -f patch.yaml
  rm patch.yaml
}

function patchAlertManagerForMTLS() {
  patch=$(cat <<"EOF"
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
  )

  echo "${patch}" > patch.yaml
  kubectl apply -f patch.yaml
  rm patch.yaml
}

function patchDeploymentsToInjectSidecar() {
  allDeploy=(
    kiali-server
    monitoring-kube-state-metrics
    monitoring-operator
  )

  resource="deployment"
  namespace="kyma-system"

  for depl in "${allDeploy[@]}"; do
    if kubectl get ${resource} -n ${namespace} "${depl}" > /dev/null; then
      kubectl get ${resource} -n ${namespace} "${depl}" -o yaml > "${depl}.yaml"

      if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' -e 's/sidecar.istio.io\/inject: "false"/sidecar.istio.io\/inject: "true"/g' "${depl}.yaml"
      else # assume Linux otherwise
        sed -i 's/sidecar.istio.io\/inject: "false"/sidecar.istio.io\/inject: "true"/g' "${depl}.yaml"
      fi

      kubectl apply -f "${depl}.yaml" || true

      rm "${depl}.yaml"
    fi
  done
}

function enableNodeExporterMTLS() {
  # The patches around the DaemonSet involve an addition of two init containers that together setup certificates
  # for the node-exporter application to use. There are also two new mounts - a shared directory (node-certs)
  # and the Istio CA secret (istio-certs).

  daemonset=$(cat <<"EOF"
apiVersion: apps/v1
kind: DaemonSet
metadata:
  annotations:
    deprecated.daemonset.template.generation: "1"
    meta.helm.sh/release-name: monitoring
    meta.helm.sh/release-namespace: kyma-system
  labels:
    app: prometheus-node-exporter
    app.kubernetes.io/instance: monitoring
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: prometheus-node-exporter
    chart: prometheus-node-exporter-1.12.0
    helm.sh/chart: prometheus-node-exporter-1.12.0
    jobLabel: node-exporter
    release: monitoring
  name: monitoring-prometheus-node-exporter
  namespace: kyma-system
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: prometheus-node-exporter
      release: monitoring
  template:
    metadata:
      labels:
        app: prometheus-node-exporter
        app.kubernetes.io/instance: monitoring
        app.kubernetes.io/managed-by: Helm
        app.kubernetes.io/name: prometheus-node-exporter
        chart: prometheus-node-exporter-1.12.0
        helm.sh/chart: prometheus-node-exporter-1.12.0
        jobLabel: node-exporter
        release: monitoring
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
        - --path.procfs=/host/proc
        - --path.sysfs=/host/sys
        - --path.rootfs=/host/root
        - --web.listen-address=$(HOST_IP):9100
        - --web.config=/etc/certs/web.yaml
        - --collector.filesystem.ignored-mount-points=^/(dev|proc|sys|var/lib/docker/.+)($|/)
        - --collector.filesystem.ignored-fs-types=^(autofs|binfmt_misc|cgroup|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|mqueue|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|sysfs|tracefs)$
        env:
        - name: HOST_IP
          value: 0.0.0.0
        image: eu.gcr.io/kyma-project/tpi/node-exporter:1.0.1-1de56388
        imagePullPolicy: IfNotPresent
        name: node-exporter
        livenessProbe: null
        readinessProbe: null
        ports:
        - containerPort: 9100
          hostPort: 9100
          name: metrics
          protocol: TCP
        resources: {}
        securityContext:
          allowPrivilegeEscalation: false
          privileged: false
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/certs
          name: node-certs
        - name: istio-certs
          mountPath: /etc/istio/certs
        - mountPath: /host/proc
          name: proc
          readOnly: true
        - mountPath: /host/sys
          name: sys
          readOnly: true
        - mountPath: /host/root
          mountPropagation: HostToContainer
          name: root
          readOnly: true
      dnsPolicy: ClusterFirst
      hostNetwork: true
      hostPID: true
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext:
        fsGroup: 65534
        runAsGroup: 65534
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccount: monitoring-prometheus-node-exporter
      serviceAccountName: monitoring-prometheus-node-exporter
      terminationGracePeriodSeconds: 30
      tolerations:
      - effect: NoSchedule
        operator: Exists
      volumes:
      - name: istio-certs
        secret:
          secretName: istio-ca-secret
      - name: node-certs
        emptyDir: {}
      - hostPath:
          path: /proc
          type: ""
        name: proc
      - hostPath:
          path: /sys
          type: ""
        name: sys
      - hostPath:
          path: /
          type: ""
        name: root
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate

EOF
  )
  echo "$daemonset" > daemonset.yaml

  kubectl get secret istio-ca-secret --namespace=istio-system -o yaml | grep -v '^\s*namespace:\s' | kubectl replace --force --namespace=kyma-system -f -

  kubectl apply -f daemonset.yaml

  rm daemonset.yaml
} 

function patchKymaServiceMonitorsForMTLS() {
  kymaSvcMonitors=(
    istio-component-monitor
    monitoring-alertmanager
    monitoring-kube-state-metrics
    monitoring-operator
    monitoring-prometheus
    monitoring-prometheus-istio-server-server
    monitoring-prometheus-node-exporter
    monitoring-prometheus-pushgateway
    tracing-metrics
    ory-oathkeeper-maester
  )

  crd="servicemonitors.monitoring.coreos.com"
  namespace="kyma-system"
  scheme="https"
  tlsConfig='{
    "caFile": "/etc/prometheus/secrets/istio.default/root-cert.pem",
    "certFile": "/etc/prometheus/secrets/istio.default/cert-chain.pem",
    "keyFile": "/etc/prometheus/secrets/istio.default/key.pem",
    "insecureSkipVerify": true
  }'

  for sm in "${kymaSvcMonitors[@]}"; do
    if kubectl get ${crd} -n ${namespace} "${sm}" > /dev/null; then
      kubectl get ${crd} -n ${namespace} "${sm}" -o json > "${sm}.json"

      cp "${sm}.json" tmp.json
      jq --arg newSchema "$scheme" '.spec.endpoints[].scheme = $newSchema' tmp.json > "${sm}.json"
      cp "${sm}.json" tmp.json
      jq --argjson newTlsConfig "$tlsConfig" '.spec.endpoints[].tlsConfig = $newTlsConfig' tmp.json > "${sm}.json"
      rm tmp.json

      kubectl apply -f "${sm}.json" || true

      rm "${sm}.json"
    fi
  done

  kubectl delete servicemonitors.monitoring.coreos.com -n kyma-system ory-hydra-maester || true
}

function removeKymaPeerAuthsForPrometheus() {
  crd="peerauthentications.security.istio.io"
  namespace="kyma-system"

  allPAs=(
    kiali
    logging-fluent-bit-metrics
    logging-loki
    monitoring-grafana-policy
    ory-oathkeeper-maester-metrics
    ory-hydra-maester-metrics
    tracing-jaeger-operator-metrics
    tracing-jaeger-metrics
  )

  for pa in "${allPAs[@]}"; do
    kubectl delete ${crd} -n ${namespace} "${pa}" || true
  done
}
