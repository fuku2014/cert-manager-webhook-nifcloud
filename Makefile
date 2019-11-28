IMAGE_NAME := "fuku2014/cert-manager-webhook-nifcloud"
IMAGE_TAG := "0.1.0"

verify:
	go test -v .

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

release: build
	docker push $(IMAGE_NAME):$(IMAGE_TAG)
