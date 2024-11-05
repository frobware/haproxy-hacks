#!/usr/bin/env bash
oc patch route mytest --type=merge -p '{"spec":{"to":{"name":"service-unsecure"}}}'
