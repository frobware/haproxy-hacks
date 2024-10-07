#!/usr/bin/env bash

set -e

endpoints="/duplicate-te-both /duplicate-te-headers /duplicate-te-trailers /healthz /single-te /single-te-with-trailer"

# Retrieve routes and iterate through each host
oc get routes --no-headers -o custom-columns=HOST:.spec.host | while read -r host; do
    if [ -n "$host" ]; then
        # Curl with -L to follow redirects, -I for headers only, -o
        # /dev/null to discard body, -sS runs silently, but will show
        # error messages if an error occurs, and -w for status code.
        # -k to ignore certificate issues.
        #
        # curl version 8.5.0 and earlier allowed duplicate
        # Transfer-Encoding headers. Later versions started rejecting
        # these headers.
        for path in $endpoints; do
            output=$(${CURL:-curl-8.5.0} --no-keepalive -L -I -sS -o /dev/null -w "%{http_code}" -k "https://${host}${path}")
            if [ $? -ne 0 ]; then
                exit $?
            fi

            read -r http_code <<< "$output"

            if [ $http_code -eq 200 ]; then
                echo "$http_code $host/$path"
            else
                echo "$http_code: **** $host/$path check FAILED ****" >&2
            fi
        done
    else
        echo "No route found."
    fi
done
