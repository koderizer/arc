VIZ_IMAGE=arcviz:dev
MINVIZ_IMAGE=marcviz:dev
VIZ_PORT=10000
VIZ_CONTAINER_PORT=10000
PLANT_UML_CONTAINER_PORT=8080
PLANT_UML_PORT=8080
PLANTUML_IMAGE=plantuml/plantuml-server:jetty

arcviz:
	docker build -t $(VIZ_IMAGE) .

minviz: 
	docker build -t $(MINVIZ_IMAGE) -f Dockerfile.arcviz .

dev:
	docker run --name plantuml -d -p $(PLANT_UML_PORT):$(PLANT_UML_CONTAINER_PORT) $(PLANTUML_IMAGE) 
	docker run --name vizdev -d -p $(VIZ_PORT):$(VIZ_CONTAINER_PORT) $(MINVIZ_IMAGE) --pumladdr http://www.plantuml.com/plantuml/

cleandev:
	docker stop vizdev && docker rm vizdev
	docker stop plantuml && docker rm plantuml

arcli:
	cd cli && make clean build
