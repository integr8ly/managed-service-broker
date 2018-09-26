TAG = 1.0.1
DOCKERORG = quay.io/integreatly
BROKER_IMAGE_NAME = managed-service-broker

.phony: build_and_push
build_and_push: build_image push


.phony: push
push:
	docker push $(DOCKERORG)/$(BROKER_IMAGE_NAME):$(TAG)

.phony: build_image
build_image: build_binary
	docker build -t $(DOCKERORG)/$(BROKER_IMAGE_NAME):$(TAG) -f ./tmp/build/broker/Dockerfile .

.phony: build_binary
build_binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./tmp/_output/bin/$(BROKER_IMAGE_NAME) ./cmd/broker

.phony: run
run:
	KUBERNETES_CONFIG=$(HOME)/.kube/config ./tmp/_output/bin/managed-service-broker --port 8080

.phony: integration
integration:
	go test ./tests/
