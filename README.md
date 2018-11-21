# Managed Service Broker

## Deploying the broker

#### Start an OpenShift cluster with the Service Catalog
```sh
#Version v3.10
$ oc cluster up --enable=service-catalog

or

$ oc cluster up
$ oc cluster add service-catalog
```

#### Building and Pushing the Docker Image
In order to build and push the image, run the following command:
```
$ make build_and_push DOCKERORG=<yourDockerOrg>

or

$ make build_image <DOCKERORG=yourDockerOrg>
$ make push <DOCKERORG=yourDockerOrg>
```

#### Deploy managed-service-broker
An OpenShift template in the `templates` directory of this repo is used to deploy the broker to a running OpenShift cluster.
This assumes that the [`svcat` command line tool](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md) is installed.

```sh
# Login as admin user
$ oc login -u system:admin

# Switch to a new project
$ oc new-project <project-name>

# Process the template and create the broker deployment
$ oc process -f templates/broker.template.yaml \
 -p NAMESPACE=<project-name> \
 -p ROUTE_SUFFIX=<clusterRouteSubDomain> \
 -p IMAGE_ORG=<yourDockerOrg> \
 -p CHE_DASHBOARD_URL=<cheDashBoardUrl> \
 -p LAUNCHER_DASHBOARD_URL=<launcherDashBoardUrl> \
 -p THREESCALE_DASHBOARD_URL=<3scaleDashBoardUrl> \
 | oc create -f -

# Verify that the broker has been registered correctly and STATUS is 'Ready'
$ svcat get brokers

# View the status of the broker
$ oc describe clusterservicebroker managed-service-broker
```

__NOTE:__ Che and Launcher do not need to be deployed to deploy the managed-service-broker, `CHE_DASHBOARD_URL` and `LAUNCHER_DASHBOARD_URL` just need to be set to a non-empty value.
#### Add syndesis-crd:

```
$ oc create -f https://raw.githubusercontent.com/syndesisio/syndesis/master/install/operator/deploy/syndesis-crd.yml
```

# Local Development (Minishift)
Guide to building and running the broker locally and connecting it to a minishift VM.

__Note:__ The same steps should work with any OpenShift cluster (oc cluster up) that has access to your host machine.

#### Start minishift VM:

OpenShift Version 3.9.0:
```
$ minishift start --openshift-version v3.9.0 --extra-clusterup-flags "--service-catalog"
```

OpenShift Version 3.10.0:
```
$ minishift start --openshift-version v3.10.0 --extra-clusterup-flags "--enable=*,service-catalog"
```

```
$ oc login -u system:admin && oc adm policy add-cluster-role-to-user cluster-admin developer && oc login -u developer -p any && minishift console
$ eval $(minishift docker-env)
```

#### Add syndesis-crd:

```
$ oc create -f https://raw.githubusercontent.com/syndesisio/syndesis/master/install/operator/deploy/syndesis-crd.yml
```

#### Setup local broker:

When setting up the broker we need to set the URL that the OpenShift cluster can access your locally running broker on. In the case of minishift this will be something like "192.168.99.1"
```
$ oc process -f templates/broker.local.template.yml -p URL=http://192.168.99.1:8080 | oc create -f -
```

Alternatively, if you already have a running managed service broker in your cluster you can patch the existing resource:
```
$ oc patch clusterservicebroker/managed-service-broker --patch '{"spec":{"url": "http://192.168.99.1:8080"}}'
```

#### Build and run the broker locally:

```bash
# Make sure KUBERNETES_CONFIG is set
# Default location is ~/.kube/config
$ export KUBERNETES_CONFIG=~/.kube/config
```

```
$ make build_binary run
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./tmp/_output/bin/managed-service-broker ./cmd/broker
KUBERNETES_CONFIG=/home/mnairn/.kube/config ./tmp/_output/bin/managed-service-broker --port 8080
INFO[0000] Catalog()
INFO[0000] Getting fuse catalog entries
INFO[0000] Getting launcher catalog entries
INFO[0000] Getting che catalog entries
INFO[0000] Starting server on :8080
```

#### Verify the broker exists:
```
$ svcat get brokers
           NAME                                                        URL                                              STATUS
+-------------------------+-------------------------------------------------------------------------------------------+--------+
  msb-local                 http://192.168.99.1:8080                                                                    Ready
  template-service-broker   https://apiserver.openshift-template-service-broker.svc:443/brokers/template.openshift.io   Ready
```

## Tests

#### Integration Tests

Setup the managed-service-broker as outlined in [Deploy managed-service-broker](#deploying-the-broker) section.
```bash
# Set env vars for managed-service-broker url and API token.

# Expose a route to the managed service broker
$ oc expose svc/msb
$ oc get route msb -o=jsonpath='{.status.ingress[0].host}'
msb-managed-service-broker.127.0.0.1.nip.io

# Get API token
$ oc whoami -t
EkoH4sIC1aTBWRBNeAzYfkMoMc36W2V3nqPigulKK-s

$ export KUBERNETES_API_TOKEN=EkoH4sIC1aTBWRBNeAzYfkMoMc36W2V3nqPigulKK-s
$ export BROKER_URL=http://msb-managed-service-broker.127.0.0.1.nip.io // Add protocol: http://....

# If KUBERNETES_CONFIG is not already set.
export KUBERNETES_CONFIG=~/.kube/config

# Run tests
$ make integration
```

__NOTE:__ If running the test against a [locally running managed-service-broker](#local-development-minishift) the permissions used will be of the user you have logged in with using `oc login`.
To correctly test permissions run against a managed-service-broker running in a cluster.

## How the broker uses TLS/SSL

When deploying to an OpenShift cluster, the broker is configured for TLS/SSL using the CA built into OpenShift. 
This is done by adding an OpenShift specific annotation to the broker's `Service` definition: 

```yaml
...
kind: Service
metadata:
  name: msb
  labels:
    app:  managed-service-broker
    service: msb
  annotations:
    service.alpha.openshift.io/serving-cert-secret-name: msb-tls
...
```

The annotation uses the built in CA to generate a signed cert and key in a secret called `msb-tls`. The certs are then added as environment variables to the broker container:

```yaml
    env:
    - name: TLS_CERT
        valueFrom:
        secretKeyRef:
            name: msb-tls
            key: tls.crt
    - name: TLS_KEY
        valueFrom:
        secretKeyRef:
            name: msb-tls
            key: tls.key
```

The Service Catalog must be provided with the caBundle so that it can validate the certificate signing chain. 
The CA is specified in the `ClusterServiceBroker` definition, in `spec.caBundle`:

```yaml
kind: ClusterServiceBroker
  metadata:
    name: managed-service-broker
  spec:
    caBundle: LS0tLS1CRUd...
```

To get the caBundle, run:
```sh
$ oc get secret -n kube-service-catalog -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | tail -n1
```

To prompt the catalog to read the broker's catalog end-point, you can use:
```
$ svcat sync broker managed-service-broker
```