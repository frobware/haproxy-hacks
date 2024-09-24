#!/bin/bash

oc get routes --no-headers -o custom-columns=HOST:.spec.host | while read -r host; do
    if [ -n "$host" ]; then
        # Curl with -L to follow redirects and -I to fetch headers -o
        # /dev/null discards body, -s is silent, -w outputs status
        # code. And -k because there's no money involved.
        if ! http_status=$(curl-8.5.0 -L -I -s -o /dev/null -w "%{http_code}" -k "https://$host"); then
            echo "$host failed." >&2
            continue
            # exit 1
        fi


        if [ "$http_status" -ge 400 ]; then
            echo "$host failed with status code $http_status" >&2
        else
            echo "$host status $http_status"
        fi
    else
        echo "No route found."
        exit 1
    fi
done
