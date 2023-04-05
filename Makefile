.PHONY: all dev examples

all:

dev:
	IMAGE_LOCAL=$(IMAGE) docker buildx bake --push

examples:
	./hack/check-example.sh $(IMAGE) ./examples/*
