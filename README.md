# ACME webhook for NIFCLOUD

### Installing

```bash
$ kubectl apply -f 
```

### Download and update example issuer and cert files

Apply updated yaml files to create a clusterissuer and a test certificate:

```bash
kubectl apply -f issuer_certificate.yaml
```
An example manifest:

```yml
apiVersion: v1
kind: Secret
metadata:
  name: nifcloud-api-key
  namespace: cert-manager
type: Opaque
data:
  accesskey-id: <base64 encoded accessKeyId>
  accesskey-secret: <base64 encoded accessKeySecret>
---
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: cluster-issuer-letsencrypt-wildcard-staging
spec:
  acme:
    email: tester@example.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-staging-account-key
    solvers:
    - dns01:
        webhook:
          groupName: acme.example.com
          solverName: nifcloud
---
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: wildcard-certificate-test
  namespace: default
spec:
  secretName: test-example-com-tls
  commonName: test.example.com
  dnsNames:
  - '*.test.example.com'
  issuerRef:
    name: cluster-issuer-letsencrypt-wildcard-staging
    kind: ClusterIssuer
  acme:
    config:
      - dns01:
          provider: dns
        domains:
          - '*.test.example.com'

```
