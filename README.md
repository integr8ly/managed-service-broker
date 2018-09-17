# Managed Services Broker

## Deploying the broker

#### Start an OpenShift cluster with the Service Catalog
```sh
#Version v3.10
oc cluster up
oc cluster add service-catalog
```

#### Deploy managed-services-broker
An OpenShift template in the `templates` directory of this repo is used to deploy the broker to a running OpenShift cluster.
This assumes that the [`svcat` command line tool](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md) is installed.

```sh
# Build and push the broker image
make build_image <DOCKERORG=yourDockerOrg>
make push <DOCKERORG=yourDockerOrg>

# Login as admin user
oc login -u system:admin

# Switch to a new project
oc new-project <project-name>

# Process the template and create the broker deployment
oc process -f templates/broker.template.yaml \
-p NAMESPACE=<project-name> \
-p ROUTE_SUFFIX=<clusterRouteSubDomain> \
-p IMAGE_ORG=<yourDockerOrg> \
-p CHE_DASHBOARD_URL=<cheDashBoardUrl> \
-p LAUNCHER_DASHBOARD_URL=<launcherDashBoardUrl> \
| oc create -f -

# Verify that the broker has been registered correctly and STATUS is 'Ready'
svcat get brokers

# View the status of the broker
oc describe clusterservicebroker managed-services-broker
```

## How the broker uses TLS/SSL

When deploying to an OpenShift cluster, the broker is configured for TLS/SSL using the CA built into OpenShift. 
This is done by adding an OpenShift specific annotation to the broker's `Service` definition: 

```yaml
...
kind: Service
metadata:
  name: msb
  labels:
    app:  managed-services-broker
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
    name: managed-services-broker
  spec:
    caBundle: LS0tLS1CRUd...
```

To get the caBundle, run:
```sh
oc get secret -n kube-service-catalog -o go-template='{{ range .items }}{{ if eq .type "kubernetes.io/service-account-token" }}{{ index .data "service-ca.crt" }}{{end}}{{"\n"}}{{end}}' | tail -n1
```

To prompt the catalog to read the broker's catalog end-point, you can use:
```
svcat sync broker managed-services-broker
```

## Building and Pushing the Docker Image
In order to build and push the image, run the following command:
```
make build_and_push DOCKERORG=<yourDockerOrg>
```
