FROM plantuml/plantuml-server:jetty
USER root

RUN mkdir -p /app/arc
COPY build/* /app/arc/
COPY start.sh /app/arc/start.sh
RUN chmod -R +wx /app/arc
WORKDIR /arc
CMD ["/app/arc/start.sh"]