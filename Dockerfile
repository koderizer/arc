FROM golang:alpine as builder
RUN mkdir /build
WORKDIR /build
COPY . .

RUN GOOS=linux go build -o arcviz viz/main.go

FROM plantuml/plantuml-server:jetty
USER root
COPY --from=builder /build/arcviz /usr/bin/viz
# COPY --from=builder /build/viz/script/vizinit_sysv.sh /etc/init.d/vizinit
COPY --from=builder /build/viz/script/viz-entrypoint.sh /viz-entrypoint.sh
# COPY --from=builder /build/viz/script/vizinit.service /etc/systemd/system/multi-user.target.wants/vizinit.service
RUN chmod 755 /usr/bin/viz
# RUN chmod 755 /etc/init.d/vizinit
RUN chmod 755 /viz-entrypoint.sh

EXPOSE 10000
RUN apt-get update && \
    apt-get install -y --no-install-recommends musl && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

ENTRYPOINT [ "/viz-entrypoint.sh" ]
USER jetty