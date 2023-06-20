.PHONY: all
all:

.PHONY: dev
dev:
	IMAGE_LOCAL=$(IMAGE) docker buildx bake --push

.PHONY: examples
examples:
	./hack/check-example.sh $(IMAGE) ./examples/*

.PHONY: vendor
vendor:
	$(eval $@_TMP_OUT := $(shell mktemp -d -t buildkit-output.XXXXXXXXXX))
	docker buildx bake --set "*.output=type=local,dest=$($@_TMP_OUT)" update-vendor
	rm -rf ./vendor
	cp -R "$($@_TMP_OUT)"/out/* .
	rm -rf "$($@_TMP_OUT)"/
