# Arc

__Software architecture made simple__

Arc is a simple utility to author, visualize, inspect and update software architecture design easily through simple YAML files. 

It is the utility to adopt the [C4model](https://c4model.com) architecture design approach that is promoted by Simon Brown. Arc allow the practice of architecture as code, allow easy integration to CiCd workflows such that archictecture information is accumulated and kept up to date as the system complexity grows, easily and free of any additional manual effort. 

The objective is to design and deliver software product with clear architecture and clean working code, with the least amount of effort and duplicating cognitive workload. 

The workflow using arc utilities is: 
 - Start note down architecture model of your app on a yaml file named arc.yaml.
 - Use arcli to inspect and arcviz service to visualize the architecture in different views such as Landscape, Context, Containers or Component.
 - Share the visualization through a centrally hosted arcviz server if your team fancy.  

## Install

- Install and run arcviz on your machine. You need  __Docker__ as it is the only option now.
```
    docker run -d -p 10000:10000 -p 8080:8080 koderizer/arcviz:latest
```

- Install arcli utility, via Homebrew on Mac as it is the only convenient option now:
```
brew tap koderizer/arc

brew install arcli
```

_Build from source or download from release package and put to your bin path is another option_

## Usage
    arcli help


## Example
One simple application 
```yaml
app: arcs
desc: "Arc is a simple utility to author, view, inspect and update software architecture design"
users:
  - name: dev
    role: "one who create software service"

internal-systems:
  - name: arc
    desc: "Enable deloper to author, inspect and version control software systems design and code."
  
    containers:
    - name: cli
      runtime: arcli-binary
      technology: golang
      desc: "local utility to parse and build arc data to and from visualizations"
  
    - name: viz
      runtime: docker-jetty
      technology: "gRPC golang, plantuml"
      desc: "render visualization of archtecture design given a arc data blob specifications"
      components:
      - name: arcviz
        desc: grpc server to structure the layout into markup for renderer
      - name: plantuml-renderer
        desc: vizualize server using plantuml

external-systems:
  - name: dev-ide
    desc: software developement editor and integrated environment

relations:
  - { s: dev, p: design and develop software, o: arc}
  - { s: arc.arcli, p: send render request (gRPC), o: arc.arcviz}
  - { s: dev-ide, p: integrate, o: arc.arcli}
```

to visualize this, simply run from the same directory this file is in:

    arcli inspect 


**This project is underconstruction**

Utilizing and base on works done from:
- [PlantUML](https://github.com/plantuml/plantuml)
- [C4-PlantUML](https://github.com/RicardoNiepel/C4-PlantUML)