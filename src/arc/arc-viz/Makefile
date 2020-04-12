PORT=8086
CONTAINER_PORT=8080

all: build

build: clean
	mkdir -p build
	cp -R templates build/
	cp -R C4-PlantUML build/
	go build -o build/arcviz main.go
	docker build -t plantuml-server .

clean:
	rm -rf build
run:
	docker run -p $(PORT):$(CONTAINER_PORT) plantuml-server