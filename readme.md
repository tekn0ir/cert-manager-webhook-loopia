# Cert-Manager ACME webhook for Loopia (cert-manager-webhook-loopia)

> [!WARNING]
> This is the personal fork of [tekn0ir](https://github.com/tekn0ir) of [Identitry/cert-manager-webhook-loopia](https://github.com/Identitry/cert-manager-webhook-loopia).
> It exists to keep the webhook working against the current Loopia API and current cert-manager and Kubernetes releases, but it is **sparsely maintained**: expect long gaps between updates and no support beyond the Loopia DNS-01 use case it was forked for.
> Pull requests and patches are welcome, but there is no promise of a quick response.

**Last updated: 2026-09-16** — modernised to Go 1.26, cert-manager v1.21.2, Kubernetes libraries v1.37 and Alpine 3.24, with image publishing moved from Docker Hub to the GitHub Container Registry.

`cert-manager-webhook-loopia` is an ACME webhook for [Cert-Manager](https://cert-manager.io/) that allows for [Cert-Manager] to use `DNS-01` challenge against the [Loopia](https://loopia.com) DNS.

[![test](https://github.com/tekn0ir/cert-manager-webhook-loopia/actions/workflows/test.yml/badge.svg)](https://github.com/tekn0ir/cert-manager-webhook-loopia/actions/workflows/test.yml)

[![release](https://github.com/tekn0ir/cert-manager-webhook-loopia/actions/workflows/release.yml/badge.svg)](https://github.com/tekn0ir/cert-manager-webhook-loopia/actions/workflows/release.yml)

## Table of Contents

1. [Overview](#1-overview)

   1.1. [Building](#11-building)

   1.2. [Docker Image](#12-docker-image)

   1.3. [Compatibility](#13-compatibility)
2. [Installation](#2-installation)

   2.2. [Install Cert-Manager](#22-install-cert-manager)

   2.3. [Install/Uninstall Loopia Webhook](#23-install/uninstall-loopia-webhook)

3. [Using the Loopia Webhook](#3-using-the-loopia-webhook)

   3.1. [Loopia API credential Secret](#31-loopia-api-credential-secret)

   3.2. [Cert-Manager Issuer configuration](#32-cert-manager-issuer-configuration)

   3.3. [Cert-Manager Certificate configuration](#33-cert-manager-certificate-configuration)

   3.4. [Troubleshooting](#34-troubleshooting)
4. [Conformance Testing](#4-conformance-testing)

## 1. Overview

[Cert-Manager](https://cert-manager.io) is a [Kubernetes](https://kubernetes.io/) certificate management controller, it allows for issuing certifaces from a variety of sources. `cert-manager-webhook-loopia` that acts as an extension to cert-manager is targeted the use of certificates issued through the [ACME-protocol](https://en.wikipedia.org/wiki/Automated_Certificate_Management_Environment) and especially the DNS01 challenge targeting the [Loopia](https://loopia.com) hosting company DNS. Issued certificates are stored in Kubernetes as [secrets](https://kubernetes.io/docs/concepts/configuration/secret) for use within Kubernetes.

The main issuer of public certificates using the ACME-protocol is [Let´s Encrypt](https://letsencrypt.org) that issues free public TLS-certificates. What´s special about Let´s Encrypt is that the lifetime of the certificates they issue are short and issuance is automated using the ACME-protocol, that means you need a way to automatically request new certificates and renew certificates when the old are about expire.

The [ACME DNS-01 challenge](https://letsencrypt.org/docs/challenge-types/#dns-01-challenge) is one of two diiferent challenges (the other is [HTTP-01](https://letsencrypt.org/docs/challenge-types/#http-01-challenge)) that you as a user of Let´s Encrypt certificates could use to prove you´re the owner of the domain the certificate is to be issued for. ACME DNS-01 challenge has an advantage over HTTP-01 in that it allows for issuance of wildcard certificates. ACME DNS-01 challenge requires you to be able to automatically add a DNS TXT record to your public DNS zone as a proof of ownership of the domain.

The role of `cert-manager-webhook-loopia` is to act as a DNS-provider and create a DNS TXT-record in the '\_acme-challenge' sub domain of the domain a certificate should be issued for, for example: '\_acme-challenge.example.com'. The value the TXT-record should contain is supplied by the ACME issuer. When the DNS01 challenge is complete, `cert-manager-webhook-loopia` is responsible for cleaning up the TXT-records created. The matching TXT-record is removed first and the whole '\_acme-challenge' sub domain is deleted once the last record in it is gone. Note that a certificate covering both the apex and the wildcard of a domain solves both of its challenges in the same '\_acme-challenge' sub domain, so the sub domain only disappears after the last of the two records has been cleaned up.

[Loopia](https://loopia.com) is a major hosting company based in Sweden but has subsidaries in Norway and Serbia but also offers services to companies and individuals in the rest of the world.

[Loopia API](https://www.loopia.com/api) that is used by `cert-manager-webhook-loopia` is an API based on XMLRPC that allows for reading and editing of your DNS domain(s) hosted at Loopia. This API becomes very handy when we need to request a lot of certificates automatically and also renew these when they expire. [Loopia-Go client](https://github.com/jonlil/loopia-go) is the client library used for communicating with Loopia API.

---
**NOTE** You need to register for special Loopia API user credentials in  [Loopia CustomerZone](https://customerzone.loopia.com/), this is also required for testing.

---

![Loopia API Logo](https://static.loopia.se/loopiaweb/images/logos/loopia-api-logo.png "Loopia API Logo")

### 1.1. Building

Build the container image `cert-manager-webhook-loopia:latest`

```shell
make build
```

### 1.2. Docker Image

Images are built and published to the GitHub Container Registry by the [release workflow](.github/workflows/release.yml):
[ghcr.io/tekn0ir/cert-manager-webhook-loopia](https://github.com/tekn0ir/cert-manager-webhook-loopia/pkgs/container/cert-manager-webhook-loopia)

The workflow pushes both `latest` and the tag of the last release on every push to `main` and on every tag, so these tags move. The Helm chart therefore defaults to the `latest` tag with `imagePullPolicy: Always`; pin both `image.tag` and the pull policy if you want a reproducible deployment.

---

**NOTE:**
GitHub Container Registry packages are private by default. Either set the visibility of the package to public, or give the webhook's service account an `imagePullSecret` for `ghcr.io`.

---

Publishing requires the repository to allow the workflow to write packages, set *Settings -> Actions -> General -> Workflow permissions* to "Read and write permissions".

Both workflows run without any credentials: they check formatting, run `go vet`, build the binary and run the tests. The conformance suite and the live Loopia API test create and delete TXT-records in a real Loopia zone, so they skip themselves unless their credentials and zone are supplied and are therefore [run locally](#4-conformance-testing) instead of in CI.

### 1.3. Compatibility

- Built with Go 1.26 using the `golang:1.27-alpine3.24` builder and the `alpine:3.24` runtime image.
- Compiled and tested against [cert-manager] v1.21.2 and the Kubernetes libraries v1.37, the conformance fixture boots a control plane of that same version.
- Container images are built for `linux/amd64`.
- Only the Loopia DNS-01 webhook API is used, its `ChallengeRequest` contract has been unchanged since cert-manager v1.2.0, so older cert-manager releases keep working, the original repository was last tested with cert-manager v1.2.0 on Kubernetes v1.20.x.

## 2. Installation

### 2.1. Prereqs

Before starting the installation of `cert-manager-webhook-loopia` the prerequisite is that you have a working Kubernetes cluster, either in the cloud or on bare metal.
You could of course use [Cert-Manager] and `cert-manager-webhook-loopia` on [Minikube](https://minikube.sigs.k8s.io/docs/), [Microk8s](https://microk8s.io/), [K3s](https://k3s.io/) or [Docker Desktop with Kubernetes enabled](https://www.docker.com/products/docker-desktop).
The installation also require that you have registered for Loopia API credentials in the [Loopia CustomerZone](https://customerzone.loopia.com), these special credentials are required for `cert-manager-webhook-loopia` to work.

### 2.2. Install Cert-Manager

The easiest way to install Cert-Manager is using Helm. For this Helm v3 or newer needs to be installed already.
This is how to install Cert-Manager using Helm, if you wish to install using manifests or using other options you can use this [instruction](https://cert-manager.io/docs/installation/kubernetes).

Add the Jetstack Helm Repository:

```shell
helm repo add jetstack https://charts.jetstack.io
```

Update Helm chart repository cache:

```shell
helm repo update
```

Install Cert-Manager (with CRD´s):

```shell
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v1.21.2 --create-namespace --set installCRDs=true
```

Verify Cert-Manager installation by getting the cert-manager running pods:

```shell
kubectl get pods --namespace cert-manager

NAME                                       READY   STATUS    RESTARTS   AGE
cert-manager-85f9bbcd97-666mx              1/1     Running   0          2m
cert-manager-cainjector-74459fcc56-r6dc8   1/1     Running   0          2m
cert-manager-webhook-57d97ccc67-jngx8      1/1     Running   0          2m
```

Note that it might take a minute or two before all pods are running.

### 2.3. Install/Uninstall Loopia Webhook

The `cert-manager-webhook-loopia` can be installed in multiple ways but the easiest is using helm:

```shell
helm repo add tekn0ir https://tekn0ir.github.io/cert-manager-webhook-loopia
helm repo update
helm install cert-manager-webhook-loopia tekn0ir/cert-manager-webhook-loopia --namespace cert-manager
```

This will install a helm chart with the pre built image available in the GitHub Container Registry as ghcr.io/tekn0ir/cert-manager-webhook-loopia.

If you wish to uninstall `cert-manager-webhook-loopia` simply run this command:

```shell
helm uninstall cert-manager-webhook-loopia --namespace cert-manager
```

## 3. Using the Loopia Webhook

Ok, now you have probably installed `cert-manager-webhook-loopia`, it´s time to configure it for getting a certificate from Let´s Encrypt.

### 3.1. Loopia API credential Secret

In order to logon to the [Loopia API](https://www.loopia.com/api) you first need a set of credentials, as a customer with Loopia you can request these in the [Loopia Customer Zone](https://customerzone.loopia.com), the usual credentials we normally use to logon with Loopia wont work.

---

**Note:**
Your Loopia API account requires these permissions:

- addZoneRecord
- getZoneRecords
- removeZoneRecord
- removeSubdomain

---

When we have the Loopia API credentials (username and password) we need to store these credentials safely within Kubernetes and Kubernetes has a special API object type, [Secret](https://kubernetes.io/docs/concepts/configuration/secret) that can be used for this.

The Secret needs to be created in the "cert-manager" namespace, otherwise permissions needs to be given for cert-manager to use the Secret.
This is the secret configuration we need to apply to Kubernetes, you can find this file in the configuration/ folder. Replace the username and password with your Loopia API credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: loopia-credentials
  namespace: cert-manager
stringData:
  username: "LOOPIA API USERNAME"
  password: "LOOPIA API PASSWORD"
```

Then deploy the Secret to Kubernetes using this command:

```shell
kubectl apply -f configuration/loopia-credentials.yaml
```

You can also deploy the secret using this command, replace the username and password with your Loopia API credentials.

```shell
kubectl create secret generic loopia-credentials --namespace cert-manager --from-literal=username='LOOPIA API USERNAME' --from-literal=password='LOOPIA API PASSWORD'
```

To remove the Secret, run this command:

```shell
kubectl delete secret loopia-credentials --namespace cert-manager
```

### 3.2 Cert-Manager Issuer configuration

Issuers and ClusterIssuers, are Kubernetes resources that represent certificate authorities (CAs) that are able to generate signed certificates. An Issuer is limited to a single namespace whereas a ClusterIssuer can issue certificates for the whole cluster.
The example yaml below is for creating a ClusterIssuer but you can just change "ClusterIssuer" to "Issuer" if you like to restrict the certificate to a single namespace.

You also need to replace the email adress to your real email adress, Let´s Encrypt needs this to identify you as a subscriber and holder of the private key. This email adress will also recieve warnings of expiring certs and notifications about changes to [Let´s Encrypts privacy policy](https://letsencrypt.org/privacy).

The ClusterIssuer example below is targeted [Let´s Encrypts staging environment](https://letsencrypt.org/docs/staging-environment), this will allow you to get things right before issuing trusted certificates and reduce the chance of your running up against rate limits.
When you have successfully tested your configuration you can remove the staging ClusterIssuer and replace it with a production one pointing to the Let´s Encrypt production environment, changing the name and the name of the Secret where the issued certificate should end up.

The example below is also available as configuration/le-staging-clusterissuer.yaml.

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL for testing.
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # The ACME server URL for production.
    # server: https://acme-v02.api.letsencrypt.org/directory

    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: hostmaster@example.com

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
      - dns01:
          webhook:
            groupName: acme.webhook.loopia.com
            solverName: loopia
            config:
              usernameSecretKeyRef:
                name: loopia-credentials
                key: username
              passwordSecretKeyRef:
                name: loopia-credentials
                key: password
```

To deploy the Cluster Issuer configuration file after you have edit it you can run this command:

```shell
kubectl apply -f configuration/le-staging-clusterissuer.yaml
```

Afterwards, check the status of the Cluster Issuer.

```shell
kubectl describe clusterissuer letsencrypt-staging
```

To delete the Cluster Issuer, run this command:

```shell
kubectl delete clusterissuer letsencrypt-staging
```

### 3.3. Cert-Manager Certificate configuration

Time to wrap this up, the final Kubernetes resource we need is a Certificate. The [Certificate resource](https://cert-manager.io/docs/usage/certificate) represents a human readable definition of a certificate request that is to be honored by an issuer which is to be kept up-to-date.

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: staging-cert-example-com
spec:
  commonName: example.com # REPLACE THIS WITH YOUR DOMAIN
  dnsNames:
  - example.com # REPLACE THIS WITH YOUR DOMAIN
  issuerRef:
    name: letsencrypt-staging
    kind: ClusterIssuer
  secretName: example-com-tls
```

To check the status of the certificate you can run this command:

```shell
kubectl describe certificate staging-cert-example-com
```

If you wish to delete the Certificate, run this command:

```shell
kubectl delete certificate staging-cert-example-com
```

### 3.4. Troubleshooting

Cert-Manager has a great page that describes how to do [troubleshooting](https://cert-manager.io/docs/faq/troubleshooting).

When a DNS-01 challenge fails with an error like:

```text
Failed to create TXT record for '_acme-challenge.example.com': Calling parameters do not match signature
```

the [Loopia API] rejected the arguments of the request. Loopia expects the signature `(username, password, domain, subdomain, ...)`, so the webhook needs [Loopia-Go client](https://github.com/jonlil/loopia-go) revision `v0.0.0-20220617090554-b07b4c905b28` or later. Older revisions send an additional customer number argument, which makes every Loopia call fail with `Fault 623`.

Since version 1.2 every error the solver returns names the Loopia API call that failed, the '\_acme-challenge' sub domain and zone it was made for and the domain the certificate is issued for, for example:

```text
Loopia API call addZoneRecord for the TXT record in "_acme-challenge.example.com" (zone "example.com", ACME challenge for "*.example.com") failed: Fault(623): Calling parameters do not match signature
```

cert-manager copies that message verbatim into `Challenge.Status.Reason`, which is surfaced in the `PresentError` event, in `Order.Status.Reason` and in the status of the `Certificate`, so this is enough to identify the failing call:

```shell
kubectl get challenges --all-namespaces
kubectl describe challenge --namespace <namespace> <challenge-name>
kubectl describe order --namespace <namespace> <order-name>
kubectl describe certificate --namespace <namespace> <certificate-name>
```

The fault itself originates from the Loopia API, if the code is not self explanatory it can be looked up in the [Loopia API documentation](https://www.loopia.com/api). Two things worth checking are whether the API user still has the [permissions listed above](#31-loopia-api-credential-secret) and whether the credentials in the `loopia-credentials` Secret are still valid. The webhook pod logs show the same errors and are still useful for anything the solver cannot know about:

```shell
kubectl logs --namespace cert-manager deployment/cert-manager-webhook-loopia --follow
```

The record is written to the zone as soon as `Present` returns, but that is not the same as the record being resolvable. Loopia's nameservers can lag well behind the API: a TXT-record was measured to become visible on `ns1.loopia.se` and `ns2.loopia.se` between 20 and 45 minutes after `addZoneRecord` returned `OK`, while other record types were served within a couple of minutes. cert-manager waits for the record to propagate before it asks the ACME issuer to validate the challenge, so a slow publication shows up as a `Present` that succeeds and a challenge that still fails with a validation error. Check what the public DNS actually answers before assuming the webhook is at fault:

```shell
dig +short TXT _acme-challenge.example.com
dig +short @ns1.loopia.se TXT _acme-challenge.example.com
```

## 4. Conformance Testing

The testing of a cert-manager weebhook is a bit special and not a typical unit or integration test, instead there´s a test-fixure supplied that build up a complete Kubernetes control plane where testing is performed. This not only requires you to download a set of test binaries but also prepare some files for testing.

The test binaries are the `etcd`, `kube-apiserver` and `kubectl` binaries that controller-runtime's [envtest](https://book.kubebuilder.io/reference/envtest) boots a control plane with, they are downloaded with [setup-envtest](https://sigs.k8s.io/controller-runtime/tools/setup-envtest) in the version matching the `k8s.io/client-go` dependency in `go.mod`.

- **testdata/scripts/fetch-test-binaries.sh:**\
  Script for downloading the test binaries into testdata/bin. Run it without arguments to get the path of the binaries, or with `--env` to get the `KUBEBUILDER_ASSETS` and `PATH` exports that make them available to `go test`.

- **testdata/loopia/config.json:**\
  This is a config file that basically informs the test fixture how to find the Kubernetes secret and keys that contains the Loopia API username and password.

- **testdata/loopia/loopia-credentials.yaml:**\
  A Kubernetes secret configuration that will be applied to the Kubernetes control plane during test. Real Loopia API credentials is required since the tests connects to Loopia creating a cert-manager-dns01-tests sub domain with a TXT-record.

- **testdata/bin:**\
  Folder location for the test-binaries, it is not checked in.

`cert-manager-webhook-loopia` has been tested for conformance, not only simple create/delete TXT-record but also in Strict/Extended mode where multiple simultaneus TXT-records are tested.

The conformance suite and the live Loopia API test are **not run in CI**, they need real credentials and modify a zone you control at Loopia. The conformance suite is behind the `conformance` build tag, so a plain `go test ./...` never compiles it and needs neither credentials nor the envtest binaries; `make check` runs the same credential-free formatting, vet and test checks as CI. To run the conformance suite, install the test binaries and point the test at a zone you control at Loopia:

```shell
make test
```

```shell
export TEST_ZONE_NAME=example.com.
export TEST_STRICT_MODE=false
export TEST_PROPAGATION_LIMIT=60m
go test -tags conformance -v .
```

- **loopia_api_test.go:**\
  A separate live test of the client path used by the solver, it creates the two TXT-records a wildcard challenge needs in one sub domain, looks them up and removes them again. It is skipped unless `LOOPIA_USERNAME`, `LOOPIA_PASSWORD` and `LOOPIA_TEST_ZONE` are set:

```shell
LOOPIA_USERNAME=... LOOPIA_PASSWORD=... LOOPIA_TEST_ZONE=example.com \
  go test -run TestLoopiaAPIAuthenticationSignature -v .
```

[ACME DNS-01 challenge]: https://letsencrypt.org/docs/challenge-types/#dns-01-challenge
[ACME documentation]: https://cert-manager.io/docs/configuration/acme
[Certificate]: https://cert-manager.io/docs/usage/certificate
[Cert-Manager]: https://cert-manager.io
[Let´s Encrypt]: https://letsencrypt.org
[Loopia]: https://loopia.com/
[Loopia Customer Zone]: https://www.loopia.com/login
[Loopia API]: https://www.loopia.com/api
[Helm]: https://helm.sh
[image tags]: https://github.com/tekn0ir/cert-manager-webhook-loopia/pkgs/container/cert-manager-webhook-loopia
[Kubernetes]: https://kubernetes.io
