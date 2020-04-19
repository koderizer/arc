PLANT_UML_IMAGE=plantuml/plantuml-server:jetty
VIZ_IMAGE=arc-viz
BINARY_VIZ_NAME=arcviz
VIZ_PORT=8086
VIZ_CONTAINER_PORT=8086
PLANT_UML_CONTAINER_PORT=8080
PLANT_UML_PORT=8080
all: clean build docker

build: 
	mkdir build
	cp -R puml/templates build/
	cp -R puml/C4-PlantUML build/
	go build -o build/$(BINARY_VIZ_NAME) main.go

docker:
	docker build -t $(VIZ_IMAGE) .

clean:
	rm -rf build
	
dep:
	docker run -d -p $(PLANT_UML_PORT):$(PLANT_UML_CONTAINER_PORT) $(PLANT_UML_IMAGE)

dev: build
	cd build && ./$(BINARY_VIZ_NAME)