FROM golang:alpine as builder
RUN mkdir /build
WORKDIR /build
COPY . .

RUN GOOS=linux go build -o arcviz viz/main.go

FROM plantuml/plantuml-server:jetty
USER root
COPY --from=builder /build/viz/puml/C4-PlantUML /C4-PlantUML
COPY --from=builder /build/arcviz /usr/bin/arcviz
COPY --from=builder /build/viz/script/viz-entrypoint.sh /viz-entrypoint.sh
RUN chmod 755 /usr/bin/arcviz
RUN chmod 755 /viz-entrypoint.sh

EXPOSE 10000
RUN apt-get update && \
    apt-get install -y --no-install-recommends musl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*
ENV ALLOW_PLANTUML_INCLUDE=true
ENTRYPOINT [ "/viz-entrypoint.sh" ]
USER jetty