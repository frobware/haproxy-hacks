#!/usr/bin/env bash

while :
do
    oc delete route bz1941703-ephemeral
    sleep 1.5
    oc expose service bz1941703 --name=bz1941703-ephemeral --port 8080
    sleep 2.5
done