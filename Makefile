.PHONY: all dev

all:

dev:
	IMAGE_LOCAL=$(IMAGE) docker buildx bake --push
