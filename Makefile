VIZ_IMAGE=arcviz
VIZ_PORT=10000
VIZ_CONTAINER_PORT=10000
PLANT_UML_CONTAINER_PORT=8080
PLANT_UML_PORT=8080

arcviz:
	docker build -t $(VIZ_IMAGE) .

dev:
	docker run --name arcvizdev -d -p $(VIZ_PORT):$(VIZ_CONTAINER_PORT) -p $(PLANT_UML_PORT):$(PLANT_UML_CONTAINER_PORT) $(VIZ_IMAGE)

cleandev:
	docker stop arcvizdev && docker rm arcvizdev

arcli:
	cd cli && make clean build
