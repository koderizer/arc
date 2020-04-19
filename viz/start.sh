#!/bin/bash
# /app/arcviz &
java -Djetty.contextpath=/ -jar $JETTY_BASE/target/dependency/jetty-runner.jar $JETTY_BASE/target/plantuml.war