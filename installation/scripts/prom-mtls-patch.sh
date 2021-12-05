function prometheusMTLSPatch() {
  patchPrometheusForMTLS
  patchAlertManagerForMTLS
  enableNodeExporterTLS
  patchDeploymentsToInjectSidecar
  patchKymaServiceMonitorsForMTLS
  removeKymaPeerAuthsForPrometheus
  patchMonitoringTests
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
    api-gateway
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

function enableNodeExporterTLS() {
  monitor=$(cat <<"EOF"
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  annotations:
    meta.helm.sh/release-name: monitoring
    meta.helm.sh/release-namespace: kyma-system
  labels:
    app: monitoring-node-exporter
    app.kubernetes.io/instance: monitoring
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: monitoring
    chart: monitoring-1.0.0
    helm.sh/chart: monitoring-1.0.0
    release: monitoring
  name: monitoring-node-exporter
  namespace: kyma-system
spec:
  endpoints:
  - metricRelabelings:
    - action: keep
      regex: ^(go_goroutines|go_memstats_alloc_bytes|go_memstats_heap_alloc_bytes|go_memstats_heap_inuse_bytes|go_memstats_heap_sys_bytes|go_memstats_stack_inuse_bytes|node_.*|process_cpu_seconds_total|process_max_fds|process_open_fds|process_resident_memory_bytes|process_start_time_seconds|process_virtual_memory_bytes)$
      sourceLabels:
      - __name__
    port: metrics
    scheme: https
    tlsConfig:
      insecureSkipVerify: true
  jobLabel: jobLabel
  selector:
    matchLabels:
      app: prometheus-node-exporter
      release: monitoring

EOF
  )
  echo "$monitor" > monitor.yaml

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
        image: frapsoft/openssl
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
        image: busybox
        command: ['sh', '-c', 'printf "tls_server_config:\n  cert_file: /etc/certs/node.crt\n  key_file: /etc/certs/node.key" > /etc/certs/web.yaml'] 
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
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /
            port: 9100
            scheme: HTTPS
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: node-exporter
        ports:
        - containerPort: 9100
          hostPort: 9100
          name: metrics
          protocol: TCP
        readinessProbe:
          failureThreshold: 3
          httpGet:
            path: /
            port: 9100
            scheme: HTTPS
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        resources: {}
        securityContext:
          allowPrivilegeEscalation: false
          privileged: false
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /etc/certs
          name: node-certs
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

  kubectl apply --force -f monitor.yaml
  kubectl apply --force -f daemonset.yaml

  rm monitor.yaml
  rm daemonset.yaml
} 

function patchKymaServiceMonitorsForMTLS() {
  kymaSvcMonitors=(
    kiali
    logging-fluent-bit
    logging-loki
    ory-oathkeeper-maester
    ory-hydra-maester
    tracing-jaeger-operator
    tracing-jaeger
    monitoring-grafana
    monitoring-alertmanager
    monitoring-operator 
    monitoring-kube-state-metrics 
    dex
    api-gateway
    monitoring-prometheus-pushgateway
  )

  crd="servicemonitors.monitoring.coreos.com"
  namespace="kyma-system"
  patchContent=$(cat <<"EOF"
    scheme: https
    tlsConfig:
      caFile: /etc/prometheus/secrets/istio.default/root-cert.pem
      certFile: /etc/prometheus/secrets/istio.default/cert-chain.pem
      keyFile: /etc/prometheus/secrets/istio.default/key.pem
      insecureSkipVerify: true

EOF
  )

  echo "$patchContent" > tmp_patch_content.yaml

  for sm in "${kymaSvcMonitors[@]}"; do
    if kubectl get ${crd} -n ${namespace} "${sm}" > /dev/null; then
      kubectl get ${crd} -n ${namespace} "${sm}" -o yaml > "${sm}.yaml"

      if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' -e '/ port:/r tmp_patch_content.yaml' "${sm}.yaml"
        sed -i '' -e '/ path:/r tmp_patch_content.yaml' "${sm}.yaml"
      else # assume Linux otherwise
        sed -i '/ port:/r tmp_patch_content.yaml' "${sm}.yaml"
        sed -i '/ path:/r tmp_patch_content.yaml' "${sm}.yaml"
      fi

      kubectl apply -f "${sm}.yaml" || true

      rm "${sm}.yaml"
    fi
  done

  rm tmp_patch_content.yaml
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
    monitoring-prometheus-pushgateway
  )

  for pa in "${allPAs[@]}"; do
    kubectl delete ${crd} -n ${namespace} "${pa}" || true
  done
}

function patchMonitoringTests() {
  crd="testdefinitions"
  namespace="kyma-system"
  name="monitoring"

  patchSidecarContainerCommand=$(cat <<"EOF"
        - until curl -fsI http://localhost:15021/healthz/ready; do echo \"Waiting
          for Sidecar...\"; sleep 3; done; echo \"Sidecar available. Running the command...\";
          ./test-monitoring; x=$(echo $?); curl -fsI -X POST http://localhost:15020/quitquitquit
          && exit $x
EOF
  )

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
