CONTAINER_NAME=hatotaka/nasne-exporter

container:
	GOOS=linux CGO_ENABLED=0 go build -o nasne-exporter
	docker build -t $(CONTAINER_NAME):local .
