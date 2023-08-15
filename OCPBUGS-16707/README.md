https://issues.redhat.com/browse/OCPBUGS-16707

https://issues.redhat.com/browse/OCPBUGS-16707
- Created three routes with the similar hosts, one without the path and other eith the paths defined.

# oc get routes
NAME     HOST/PORT                                                                       PATH    SERVICES        PORT   TERMINATION   WILDCARD
route1   httpd-example-path-based-routes.apps.firstcluster.lab.upshift.rdu2.redhat.com           httpd-example   web    edge          None
route2   httpd-example-path-based-routes.apps.firstcluster.lab.upshift.rdu2.redhat.com   /path   httpd-example   web    edge          None
route3   HostAlreadyClaimed                                                              /path   httpd-example   web    edge          None   <---------------


- Got 'HostAlreadyClaimed' error for the third route 'route3' which is
  expected because the path and the hostname of 'route2' & route3' are
  the same.

- In the route description, we could see that the first route that is
  'route1' is reported to be the older route for the host but we
  expect it should report 'route2' because the hostname and paths are
  similar for the route2 and route3.

# oc describe route route3
Name:            route3
Namespace:        path-based-routes
Created:        14 seconds ago
Labels:            app=httpd-example
            template=httpd-example
Annotations:        <none>
Requested Host:        httpd-example-path-based-routes.apps.firstcluster.lab.upshift.rdu2.redhat.com
            rejected by router default:  (host router-default.apps.firstcluster.lab.upshift.rdu2.redhat.com)HostAlreadyClaimed (14 seconds ago)
              route route1 already exposes httpd-example-path-based-routes.apps.firstcluster.lab.upshift.rdu2.redhat.com and is older   <----------------
Path:            /path
TLS Termination:    edge
Insecure Policy:    <none>
Endpoint Port:        web

Service:    httpd-example
Weight:        100 (100%)
Endpoints:    10.1.2.3:8080 

- However, deleting the 'route2' resolves the issue.
