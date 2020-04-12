#!/bin/bash
/arcapp/arcviz &
java -Djetty.contextpath=/ -jar target/dependency/jetty-runner.jar target/plantuml.war