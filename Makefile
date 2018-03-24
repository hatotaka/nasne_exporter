CONTAINER_NAME=hatotaka/nasne-exporter

container:
	docker build -t $(CONTAINER_NAME):local .
