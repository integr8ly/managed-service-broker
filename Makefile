TAG = 1.0.0
DOCKERORG = aerogear
BROKER_IMAGE_NAME = managed-services-broker

.phony: push_broker
push_broker:
	docker push $(DOCKERORG)/$(BROKER_IMAGE_NAME):$(TAG)

.phony: build_image
build_image: build_binary
	docker build -t $(DOCKERORG)/$(BROKER_IMAGE_NAME):$(TAG) -f ./tmp/build/broker/Dockerfile .

.phony: build_binary
build_binary:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./tmp/_output/bin/$(BROKER_IMAGE_NAME) ./cmd/broker