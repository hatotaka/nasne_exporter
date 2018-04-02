CONTAINER_NAME=quay.io/hatotaka/nasne-exporter

container:
	docker build \
		-t $(CONTAINER_NAME):local \
		-f build/package/Dockerfile \
		.
