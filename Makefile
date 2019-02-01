SHELL = /bin/bash
REG = quay.io
ORG = integreatly
IMAGE = managed-service-broker
TAG = master
PROJECT = managed-service-broker

.PHONY: code/run
code/run:
	@KUBERNETES_CONFIG=$(HOME)/.kube/config ./tmp/_output/bin/managed-service-broker --port 8080

.PHONY: code/compile
code/compile:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./tmp/_output/bin/$(IMAGE) ./cmd/broker

.PHONY: code/check
code/check:
	@diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

.PHONY: code/fix
code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: image/build
image/build: code/compile
	@docker build -t $(REG)/$(ORG)/$(IMAGE):$(TAG) -f ./tmp/build/broker/Dockerfile .

.PHONY: image/push
image/push:
	@docker push $(REG)/$(ORG)/$(IMAGE):$(TAG)

.PHONY: image/build/push
image/build/push: image/build image/push

.PHONY: test/e2e
test/e2e:
	@go test ./tests/

.PHONY: cluster/prepare
cluster/prepare:
	@oc new-project $(PROJECT)
	@oc create -f https://raw.githubusercontent.com/syndesisio/fuse-online-install/1.5/resources/fuse-online-image-streams.yml -n openshift
	@oc create -f https://raw.githubusercontent.com/integr8ly/managed-service-controller/managed-service-controller-v1.0.0/deploy/fuse-image-stream.yaml -n openshift
	@oc create -f https://raw.githubusercontent.com/syndesisio/fuse-online-install/1.5/resources/syndesis-crd.yml

.PHONY: cluster/deploy
cluster/deploy:
	@oc process -f ./templates/broker.template.yaml \
      -p IMAGE_TAG=$(TAG) \
      -p NAMESPACE=$(PROJECT) \
      -p ROUTE_SUFFIX=127.0.0.1.nip.io  \
      -p IMAGE_ORG=$(REG)/$(ORG) \
      -p CHE_DASHBOARD_URL=http://che \
      -p LAUNCHER_DASHBOARD_URL=http://launcher \
      -p THREESCALE_DASHBOARD_URL=http://3scale \
      -p APICURIO_DASHBOARD_URL=http://apicurio \
      -p MONITORING_KEY=middleware \
      | oc create -f -

.PHONY: cluster/remove/deploy
cluster/remove/deploy:
	@oc process -f ./templates/broker.template.yaml \
      -p IMAGE_TAG=$(TAG) \
      -p NAMESPACE=$(PROJECT) \
      -p ROUTE_SUFFIX=127.0.0.1.nip.io  \
      -p IMAGE_ORG=$(REG)/$(ORG) \
      -p CHE_DASHBOARD_URL=http://che \
      -p LAUNCHER_DASHBOARD_URL=http://launcher \
      -p THREESCALE_DASHBOARD_URL=http://3scale \
      -p APICURIO_DASHBOARD_URL=http://apicurio \
      -p MONITORING_KEY=middleware \
      | oc delete -f -

.PHONY: cluster/clean
cluster/clean:
	@oc delete -f https://raw.githubusercontent.com/syndesisio/fuse-online-install/1.5/resources/fuse-online-image-streams.yml -n openshift
	@oc delete -f https://raw.githubusercontent.com/integr8ly/managed-service-controller/managed-service-controller-v1.0.0/deploy/fuse-image-stream.yaml -n openshift
	@oc delete -f https://raw.githubusercontent.com/syndesisio/fuse-online-install/1.5/resources/syndesis-crd.yml
	@oc delete namespace $(PROJECT)
