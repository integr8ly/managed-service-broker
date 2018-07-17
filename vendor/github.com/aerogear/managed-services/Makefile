TAG = 1.0.0
DOCKERORG = aerogear
OPERATOR_IMAGE_NAME = managed-service-operator
BROKER_IMAGE_NAME = managed-services-broker

.phony: push_broker
push_broker:
	docker push $(DOCKERORG)/$(BROKER_IMAGE_NAME):$(TAG)

.phony: build_all
build: build_operator_image build_broker_image

.phony: build_operator_image
build_operator_image: build_operator_binary
	operator-sdk build $(DOCKERORG)/$(OPERATOR_IMAGE_NAME):$(TAG)

.phony: build_broker_image
build_broker_image: build_broker_binary
	docker build -t $(DOCKERORG)/$(BROKER_IMAGE_NAME):$(TAG) -f ./tmp/build/broker/Dockerfile .

.phony: build_operator_binary
build_operator_binary:
	env GOOS=linux GOARCH=amd64 go build -o ./cmd/operator/operator ./cmd/operator

.phony: build_broker_binary
build_broker_binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./tmp/_output/bin/$(BROKER_IMAGE_NAME) ./cmd/broker