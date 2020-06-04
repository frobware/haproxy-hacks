# BZ1841454 - debug scripts

I created some scripts to try and reproduce:

  https://bugzilla.redhat.com/show_bug.cgi?id=1841454

The general gist is that we run a server in a pod that will echo the
headers and the payload back to the client. The curl client sends
images.json as a application/json payload.
  
## Setup/Deploy

Deploy the server:

    $ ./apply

## Run jobs against the server in parallel

The following will run 1000 curls (with up to 200 curls running in
parallel) against the host for the bz1841454-edge route (the hostname
is inferred). Each invocation of curl is expected to complete within
10s.

    $ TIMEOUT=10 N=1000 P=200 ./run-jobs

## Analayzing the results

    $ N=1000 ./analyze-results

    count(curl requests that exceeded the timeout):
    162

    count(curl requests that didn't complete):
    0

    count(failed requests): 
    162

    count(successful requests): 
    838

    execution time (seconds) for successful requests: 
    min(JobRuntime) max(JobRuntime) round(avg(JobRuntime),3)
    0.680           10.003          4.724
