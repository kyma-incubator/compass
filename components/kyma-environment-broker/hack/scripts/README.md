## Create Network and CloudNAT

```bash
./k8s-gke-create-cloud-router.sh
```

## Install Kyma

1. Install k8s cluster
```bash
./k8s-gke-create-cluster-private.sh
```

>**NOTE:** If network already exist set `CREATE_NETWORK` to `false` in `vars.sh` file

2. Create certificates and inject them to ConfigMap
```bash
./kyma-lite/kyma-before-install-setup-dns.sh
```

>**NOTE:** If certificates already exist set `CREATE_CERT` to `false` in `vars.sh` file

3. Install Kyma
```bash
./kyma-lite/kyma-install.sh
```

4. Create DNS entries
```bash
./kyma-lite/kyma-after-install-set-dns.sh
```

5. Edit gateway resource
```bash
kubectl edit gateways.networking.istio.io -n kyma-system
```
remove `spec.servers.hosts` entites with name:
- `https-app-connector`
- `https-compass-mtls`

The final `Gateway` resource should looks like this:
```yaml
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  (...)
spec:
  selector:
    istio: ingressgateway
  servers:
  - hosts:
    - '*.cis.gophers.kyma.pro'
    port:
      name: https
      number: 443
      protocol: HTTPS
    tls:
      credentialName: kyma-gateway-certs
      mode: SIMPLE
  - hosts:
    - '*.cis.gophers.kyma.pro'
    port:
      name: http
      number: 80
      protocol: HTTP
    tls:
      httpsRedirect: true
```
