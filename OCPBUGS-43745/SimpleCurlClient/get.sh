#!/usr/bin/env bash

mvn install
mvn exec:java -Dexec.mainClass="com.example.SimpleCurlClient" -Dexec.args="$@"
