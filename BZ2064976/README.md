https://bugzilla.redhat.com/show_bug.cgi?id=2064976

# Run the canary locally:

```console
$ PORT=9090 SECOND_PORT=9091 ./ingress-operator serve-healthcheck
serving on 9091
serving on 9090
Serving canary healthcheck request
```

I don't see evidence from lsof and ss that connections are not getting
closed.

$ oc version
Client Version: 4.10.5
Server Version: 4.11.0-0.nightly-2022-03-16-103946
Kubernetes Version: v1.23.3+d67dd91

Make the openshift-router deployment privileged so we can run some
tools:

$ oc scale --replicas 0 -n openshift-cluster-version deployments/cluster-version-operator
$ oc scale --replicas 0 -n openshift-ingress-operator deployments ingress-operator

$ oc patch clusterversions/version --type=json --patch='[{"op":"add","path":"/spec/overrides","value":[{"kind":"Deployment","group":"apps/v1","name":"ingress-operator","namespace":"openshift-ingress-operator","unmanaged":true}]}]'
$ oc scale --replicas 0 -n openshift-ingress-operator deployments ingress-operator
$ oc patch clusterrole/openshift-ingress-router --type=strategic --patch='{"rules":[{"apiGroups":[""],"resources":["endpoints","namespaces","services"],"verbs":["list","watch"]},{"apiGroups":["authentication.k8s.io"],"resources":["tokenreviews"],"verbs":["create"]},{"apiGroups":["authorization.k8s.io"],"resources":["subjectaccessreviews"],"verbs":["create"]},{"apiGroups":["route.openshift.io"],"resources":["routes"],"verbs":["list","watch"]},{"apiGroups":["route.openshift.io"],"resources":["routes/status"],"verbs":["update"]},{"apiGroups":["security.openshift.io"],"resourceNames":["privileged"],"resources":["securitycontextconstraints"],"verbs":["use"]},{"apiGroups":["discovery.k8s.io"],"resources":["endpointslices"],"verbs":["list","watch"]}]}'
$ oc patch -n openshift-ingress deployment/router-default --patch='{"spec":{"template":{"spec":{"securityContext":{"runAsUser":0}}}}}'

Scale to one replica for easier debugging

$ oc scale -n openshift-ingress-operator ingresscontroller/default  --replicas=1

Verify that we only have one router pod

$ oc get pods -n openshift-ingress
NAME                                 READY   STATUS        RESTARTS   AGE
router-default-bcd788f9-2p5zj        2/2     Running       0          5s
router-default-d44d4fddd-m88v5       2/2     Terminating   0          147m

Use lsof to monitor established/open connections:

$ oc rsh router-default-bcd788f9-2p5zj 
Defaulted container "router" out of: router, logs
sh-4.4# ps -ef
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 11:50 ?        00:00:00 /usr/bin/openshift-router --v=2
root          32       1  0 11:50 ?        00:00:00 /usr/sbin/haproxy -f /var/lib/haproxy/conf/haproxy.config -p /var/lib/haproxy/run/haproxy.pid -x /var/lib/haproxy/run/haproxy.sock -sf 21
root          39       0  0 11:51 pts/0    00:00:00 /bin/sh
root          44      39  0 11:52 pts/0    00:00:00 ps -ef

sh-4.4# lsof -p 32
COMMAND PID USER   FD      TYPE             DEVICE SIZE/OFF      NODE NAME
haproxy  32 root  cwd       DIR              0,601     4096 127954449 /var/lib/haproxy/conf
haproxy  32 root  rtd       DIR              0,601       39 201331067 /
haproxy  32 root  txt       REG              0,601  2546664   8400718 /usr/sbin/haproxy
haproxy  32 root  mem       REG              252,4            8400718 /usr/sbin/haproxy (path dev=0,601)
haproxy  32 root  DEL       REG                0,1           14883247 /dev/zero
haproxy  32 root  mem       REG              252,4          127926420 /usr/lib64/libgcc_s-8-20200928.so.1 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127927128 /usr/lib64/libc-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951809 /usr/lib64/libpcre.so.1.2.10 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127927141 /usr/lib64/libpcreposix.so.0.0.6 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127952570 /usr/lib64/libcrypto.so.1.1.1g (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951466 /usr/lib64/libssl.so.1.1.1g (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951845 /usr/lib64/libpthread-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928166 /usr/lib64/librt-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951435 /usr/lib64/libdl-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928156 /usr/lib64/libz.so.1.2.11 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928149 /usr/lib64/libcrypt.so.1.1.0 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127926451 /usr/lib64/ld-2.28.so (path dev=0,601)
haproxy  32 root    0u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    1u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    2u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    3u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root    4u     unix 0x0000000000000000      0t0  14883251 type=DGRAM
haproxy  32 root    5u     unix 0x0000000000000000      0t0  14879703 /var/lib/haproxy/run/haproxy.sock.19.tmp type=STREAM
haproxy  32 root    6u     IPv4           14879706      0t0       TCP *:http (LISTEN)
haproxy  32 root    7u     IPv4           14879708      0t0       TCP *:https (LISTEN)
haproxy  32 root    8u     unix 0x0000000000000000      0t0  14879709 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root    9u     unix 0x0000000000000000      0t0  14879710 /var/lib/haproxy/run/haproxy-no-sni.sock.19.tmp type=STREAM
haproxy  32 root   10r     FIFO               0,13      0t0  14885925 pipe
haproxy  32 root   11w     FIFO               0,13      0t0  14885925 pipe
haproxy  32 root   12u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   13r     FIFO               0,13      0t0  14884953 pipe
haproxy  32 root   14w     FIFO               0,13      0t0  14884953 pipe
haproxy  32 root   15r     FIFO               0,13      0t0  14885926 pipe
haproxy  32 root   16w     FIFO               0,13      0t0  14885926 pipe
haproxy  32 root   17u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   18r     FIFO               0,13      0t0  14884107 pipe
haproxy  32 root   19w     FIFO               0,13      0t0  14884107 pipe
haproxy  32 root   20u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   22u     unix 0x0000000000000000      0t0  14885999 type=DGRAM
haproxy  32 root   23u     unix 0x0000000000000000      0t0  14884315 type=DGRAM
haproxy  32 root   24u     unix 0x0000000000000000      0t0  14886722 type=DGRAM
haproxy  32 root   28u     IPv4           14907283      0t0       TCP master-1.ocp411.int.frobware.com:41880->10-128-0-60.ingress-canary.openshift-ingress-canary.svc.cluster.local:webcache (ESTABLISHED)
haproxy  32 root   31u     IPv4           14906064      0t0       TCP 192-168-7-201.kubernetes.default.svc.cluster.local:https->api.ocp411.int.frobware.com:44066 (ESTABLISHED)
haproxy  32 root   32u     IPv4           14906066      0t0       TCP master-1.ocp411.int.frobware.com:52756->10-130-0-18.oauth-openshift.openshift-authentication.svc.cluster.local:sun-sr-https (ESTABLISHED)
haproxy  32 root   33u     IPv4           14904444      0t0       TCP 192-168-7-201.kubernetes.default.svc.cluster.local:https->api.ocp411.int.frobware.com:43606 (ESTABLISHED)
haproxy  32 root   34u     IPv4           14902802      0t0       TCP master-1.ocp411.int.frobware.com:52174->10-130-0-18.oauth-openshift.openshift-authentication.svc.cluster.local:sun-sr-https (ESTABLISHED)

Nothing much here. Let's rebuild openshift-ingress-operator to enable
keepalives when probing the canary route and make checks every 500ms.

modified   pkg/operator/controller/canary/controller.go
@@ -36,7 +36,7 @@ import (
 const (
 	canaryControllerName = "canary_controller"
 	// canaryCheckFrequency is how long to wait in between canary checks.
-	canaryCheckFrequency = 1 * time.Minute
+	canaryCheckFrequency = 500 * time.Millisecond
 	// canaryCheckCycleCount is how many successful canary checks should be observed
 	// before rotating the canary endpoint.
 	canaryCheckCycleCount = 5
modified   pkg/operator/controller/canary/http.go
@@ -61,7 +61,7 @@ func probeRouteEndpoint(route *routev1.Route) error {
 			// pod's environment.
 			Proxy:             http.ProxyFromEnvironment,
 			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
-			DisableKeepAlives: true, // BZ#2037447
+			DisableKeepAlives: false, // BZ#2037447
 		},
 	}
 	response, err := client.Do(request)

Rebuild and running:

$ make
hack/update-generated-crd.sh
hack/update-profile-manifests.sh
hack/update-generated-bindata.sh
CGO_ENABLED=0 GO111MODULE=on GOFLAGS=-mod=vendor go build -o ingress-operator  github.com/openshift/cluster-ingress-operator/cmd/ingress-operator

$ ENABLE_CANARY=1 ./hack/run-local.sh 

$ oc -n openshift-ingress rsh router-default-bcd788f9-2p5zj ps -ef
Defaulted container "router" out of: router, logs
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 11:50 ?        00:00:00 /usr/bin/openshift-router --
root          32       1  0 11:50 ?        00:00:03 /usr/sbin/haproxy -f /var/li
root          39       0  0 11:51 pts/0    00:00:00 /bin/sh
root          74       0  0 12:05 pts/1    00:00:00 ps -ef

COMMAND PID USER   FD      TYPE             DEVICE SIZE/OFF      NODE NAME
haproxy  32 root  cwd       DIR              0,601     4096 127954449 /var/lib/haproxy/conf
haproxy  32 root  rtd       DIR              0,601       62 201331067 /
haproxy  32 root  txt       REG              0,601  2546664   8400718 /usr/sbin/haproxy
haproxy  32 root  mem       REG              252,4            8400718 /usr/sbin/haproxy (path dev=0,601)
haproxy  32 root  DEL       REG                0,1           14883247 /dev/zero
haproxy  32 root  mem       REG              252,4          127926420 /usr/lib64/libgcc_s-8-20200928.so.1 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127927128 /usr/lib64/libc-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951809 /usr/lib64/libpcre.so.1.2.10 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127927141 /usr/lib64/libpcreposix.so.0.0.6 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127952570 /usr/lib64/libcrypto.so.1.1.1g (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951466 /usr/lib64/libssl.so.1.1.1g (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951845 /usr/lib64/libpthread-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928166 /usr/lib64/librt-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951435 /usr/lib64/libdl-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928156 /usr/lib64/libz.so.1.2.11 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928149 /usr/lib64/libcrypt.so.1.1.0 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127926451 /usr/lib64/ld-2.28.so (path dev=0,601)
haproxy  32 root    0u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    1u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    2u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    3u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root    4u     unix 0x0000000000000000      0t0  14883251 type=DGRAM
haproxy  32 root    5u     unix 0x0000000000000000      0t0  14879703 /var/lib/haproxy/run/haproxy.sock.19.tmp type=STREAM
haproxy  32 root    6u     IPv4           14879706      0t0       TCP *:80 (LISTEN)
haproxy  32 root    7u     IPv4           14879708      0t0       TCP *:443 (LISTEN)
haproxy  32 root    8u     unix 0x0000000000000000      0t0  14879709 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root    9u     unix 0x0000000000000000      0t0  14879710 /var/lib/haproxy/run/haproxy-no-sni.sock.19.tmp type=STREAM
haproxy  32 root   10r     FIFO               0,13      0t0  14885925 pipe
haproxy  32 root   11w     FIFO               0,13      0t0  14885925 pipe
haproxy  32 root   12u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   13r     FIFO               0,13      0t0  14884953 pipe
haproxy  32 root   14w     FIFO               0,13      0t0  14884953 pipe
haproxy  32 root   15r     FIFO               0,13      0t0  14885926 pipe
haproxy  32 root   16w     FIFO               0,13      0t0  14885926 pipe
haproxy  32 root   17u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   18r     FIFO               0,13      0t0  14884107 pipe
haproxy  32 root   19w     FIFO               0,13      0t0  14884107 pipe
haproxy  32 root   20u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   21u     IPv4           15073135      0t0       TCP 192.168.7.201:443->192.168.7.203:39778 (ESTABLISHED)
haproxy  32 root   22u     unix 0x0000000000000000      0t0  14885999 type=DGRAM
haproxy  32 root   23u     unix 0x0000000000000000      0t0  14884315 type=DGRAM
haproxy  32 root   24u     unix 0x0000000000000000      0t0  14886722 type=DGRAM
haproxy  32 root   25u     unix 0x0000000000000000      0t0  15074270 type=STREAM
haproxy  32 root   26u     unix 0x0000000000000000      0t0  15069084 type=STREAM
haproxy  32 root   27u     unix 0x0000000000000000      0t0  15073145 type=STREAM
haproxy  32 root   28u     IPv4           15072213      0t0       TCP 192.168.7.201:443->192.168.7.203:39794 (ESTABLISHED)
haproxy  32 root   29u     unix 0x0000000000000000      0t0  15072215 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   30u     IPv4           15047608      0t0       TCP 10.129.0.1:45054->10.130.0.18:6443 (ESTABLISHED)
haproxy  32 root   31u     IPv4           15048878      0t0       TCP 192.168.7.201:443->192.168.7.203:35820 (ESTABLISHED)
haproxy  32 root   32u     IPv4           15067994      0t0       TCP 192.168.7.201:443->192.168.7.203:39046 (ESTABLISHED)
haproxy  32 root   33u     unix 0x0000000000000000      0t0  15070580 type=STREAM
haproxy  32 root   34u     IPv4           15071352      0t0       TCP 192.168.7.201:443->192.168.7.203:39222 (ESTABLISHED)
haproxy  32 root   35u     IPv4           15068065      0t0       TCP 192.168.7.201:443->192.168.7.203:39074 (ESTABLISHED)
haproxy  32 root   36u     unix 0x0000000000000000      0t0  15075361 type=STREAM
haproxy  32 root   37u     IPv4           15073174      0t0       TCP 192.168.7.201:443->192.168.7.203:39818 (ESTABLISHED)
haproxy  32 root   38u     unix 0x0000000000000000      0t0  15073176 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   39u     unix 0x0000000000000000      0t0  15068114 type=STREAM
haproxy  32 root   40u     unix 0x0000000000000000      0t0  15072248 type=STREAM
haproxy  32 root   41u     IPv4           15075380      0t0       TCP 10.129.0.1:40746->10.130.0.63:8080 (ESTABLISHED)
haproxy  32 root   42u     IPv4           15073188      0t0       TCP 192.168.7.201:443->192.168.7.203:39822 (ESTABLISHED)
haproxy  32 root   43u     IPv4           15068054      0t0       TCP 192.168.7.201:443->192.168.7.203:39072 (ESTABLISHED)
haproxy  32 root   44u     IPv4           15075431      0t0       TCP 192.168.7.201:443->192.168.7.203:39848 (ESTABLISHED)
haproxy  32 root   45u     unix 0x0000000000000000      0t0  15076383 type=STREAM
haproxy  32 root   46u     unix 0x0000000000000000      0t0  15075433 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   47u     unix 0x0000000000000000      0t0  15069102 type=STREAM
haproxy  32 root   48u     unix 0x0000000000000000      0t0  15076418 type=STREAM
haproxy  32 root   49u     unix 0x0000000000000000      0t0  15075448 type=STREAM
haproxy  32 root   50u     IPv4           15068112      0t0       TCP 192.168.7.201:443->192.168.7.203:39100 (ESTABLISHED)
haproxy  32 root   51u     IPv4           15073252      0t0       TCP 192.168.7.201:443->192.168.7.203:39850 (ESTABLISHED)
haproxy  32 root   52u     IPv4           15073258      0t0       TCP 192.168.7.201:443->192.168.7.203:39876 (ESTABLISHED)
haproxy  32 root   53u     unix 0x0000000000000000      0t0  15077465 type=STREAM
haproxy  32 root   54u     unix 0x0000000000000000      0t0  15069053 type=STREAM
haproxy  32 root   55u     unix 0x0000000000000000      0t0  15068056 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   56u     IPv4           15077463      0t0       TCP 192.168.7.201:443->192.168.7.203:39878 (ESTABLISHED)
haproxy  32 root   57u     unix 0x0000000000000000      0t0  15075555 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   58u     IPv4           15077466      0t0       TCP 10.129.0.1:37902->10.128.0.60:8080 (ESTABLISHED)
haproxy  32 root   59u     unix 0x0000000000000000      0t0  15068067 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   60u     IPv4           15076534      0t0       TCP 192.168.7.201:443->192.168.7.203:39908 (ESTABLISHED)
haproxy  32 root   61u     unix 0x0000000000000000      0t0  15077487 type=STREAM
haproxy  32 root   62u     IPv4           15074758      0t0       TCP 10.129.0.1:44788->10.129.0.195:8080 (ESTABLISHED)
haproxy  32 root   63u     unix 0x0000000000000000      0t0  15067996 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   64u     unix 0x0000000000000000      0t0  15071291 type=STREAM
haproxy  32 root   65u     IPv4           15070030      0t0       TCP 192.168.7.201:443->192.168.7.203:39194 (ESTABLISHED)
haproxy  32 root   66u     unix 0x0000000000000000      0t0  15069912 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   67u     unix 0x0000000000000000      0t0  15070095 type=STREAM
haproxy  32 root   68u     IPv4           15069929      0t0       TCP 192.168.7.201:443->192.168.7.203:39102 (ESTABLISHED)
haproxy  32 root   69u     unix 0x0000000000000000      0t0  15070514 type=STREAM
haproxy  32 root   70u     unix 0x0000000000000000      0t0  15071354 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   71u     unix 0x0000000000000000      0t0  15069931 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   72u     IPv4           15068128      0t0       TCP 192.168.7.201:443->192.168.7.203:39128 (ESTABLISHED)
haproxy  32 root   73u     unix 0x0000000000000000      0t0  15069945 type=STREAM
haproxy  32 root   74u     unix 0x0000000000000000      0t0  15068130 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   75u     IPv4           15071368      0t0       TCP 192.168.7.201:443->192.168.7.203:39246 (ESTABLISHED)
haproxy  32 root   76u     unix 0x0000000000000000      0t0  15069167 type=STREAM
haproxy  32 root   77u     unix 0x0000000000000000      0t0  15069959 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   78u     IPv4           15069957      0t0       TCP 192.168.7.201:443->192.168.7.203:39130 (ESTABLISHED)
haproxy  32 root   79u     IPv4           15070570      0t0       TCP 192.168.7.201:443->192.168.7.203:39216 (ESTABLISHED)
haproxy  32 root   80u     IPv4           15069970      0t0       TCP 192.168.7.201:443->192.168.7.203:39156 (ESTABLISHED)
haproxy  32 root   81u     unix 0x0000000000000000      0t0  15069179 type=STREAM
haproxy  32 root   82u     IPv4           15073552      0t0       TCP 192.168.7.201:443->192.168.7.203:39370 (ESTABLISHED)
haproxy  32 root   83u     unix 0x0000000000000000      0t0  15069972 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   84u     IPv4           15069997      0t0       TCP 192.168.7.201:443->192.168.7.203:39164 (ESTABLISHED)
haproxy  32 root   85u     unix 0x0000000000000000      0t0  15071246 type=STREAM
haproxy  32 root   86u     unix 0x0000000000000000      0t0  15070605 type=STREAM
haproxy  32 root   87u     IPv4           15070015      0t0       TCP 192.168.7.201:443->192.168.7.203:39186 (ESTABLISHED)
haproxy  32 root   88u     unix 0x0000000000000000      0t0  15070535 type=STREAM
haproxy  32 root   89u     unix 0x0000000000000000      0t0  15070536 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   90u     unix 0x0000000000000000      0t0  15070032 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   91u     unix 0x0000000000000000      0t0  15071337 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   92u     unix 0x0000000000000000      0t0  15072512 type=STREAM
haproxy  32 root   93u     unix 0x0000000000000000      0t0  15073499 type=STREAM
haproxy  32 root   94u     unix 0x0000000000000000      0t0  15070156 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   95u     IPv4           15071582      0t0       TCP 192.168.7.201:443->192.168.7.203:39342 (ESTABLISHED)
haproxy  32 root   96u     unix 0x0000000000000000      0t0  15071584 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root   97u     unix 0x0000000000000000      0t0  15073017 type=STREAM
haproxy  32 root   98u     unix 0x0000000000000000      0t0  15071594 type=STREAM
haproxy  32 root   99u     unix 0x0000000000000000      0t0  15071066 type=STREAM
haproxy  32 root  100u     unix 0x0000000000000000      0t0  15073299 type=STREAM
haproxy  32 root  101u     IPv4           15070642      0t0       TCP 192.168.7.201:443->192.168.7.203:39256 (ESTABLISHED)
haproxy  32 root  102u     unix 0x0000000000000000      0t0  15071413 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  103u     unix 0x0000000000000000      0t0  15071621 type=STREAM
haproxy  32 root  104u     IPv4           15073306      0t0       TCP 192.168.7.201:443->192.168.7.203:39274 (ESTABLISHED)
haproxy  32 root  105u     unix 0x0000000000000000      0t0  15070649 type=STREAM
haproxy  32 root  106u     unix 0x0000000000000000      0t0  15073308 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  107u     unix 0x0000000000000000      0t0  15070880 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  108u     unix 0x0000000000000000      0t0  15070667 type=STREAM
haproxy  32 root  109u     IPv4           15072418      0t0       TCP 192.168.7.201:443->192.168.7.203:39284 (ESTABLISHED)
haproxy  32 root  110u     unix 0x0000000000000000      0t0  15070668 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  111u     IPv4           15070878      0t0       TCP 192.168.7.201:443->192.168.7.203:39358 (ESTABLISHED)
haproxy  32 root  112u     IPv4           15072423      0t0       TCP 192.168.7.201:443->192.168.7.203:39302 (ESTABLISHED)
haproxy  32 root  113u     unix 0x0000000000000000      0t0  15070679 type=STREAM
haproxy  32 root  114u     unix 0x0000000000000000      0t0  15073337 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  115u     unix 0x0000000000000000      0t0  15073554 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  116u     unix 0x0000000000000000      0t0  15072506 type=STREAM
haproxy  32 root  117u     IPv4           15073411      0t0       TCP 192.168.7.201:443->192.168.7.203:39312 (ESTABLISHED)
haproxy  32 root  118u     unix 0x0000000000000000      0t0  15073413 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  119u     IPv4           15071777      0t0       TCP 192.168.7.201:443->192.168.7.203:39506 (ESTABLISHED)
haproxy  32 root  120u     IPv4           15072510      0t0       TCP 192.168.7.201:443->192.168.7.203:39330 (ESTABLISHED)
haproxy  32 root  121u     unix 0x0000000000000000      0t0  15070798 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  122u     IPv4           15072639      0t0       TCP 192.168.7.201:443->192.168.7.203:39478 (ESTABLISHED)
haproxy  32 root  123u     unix 0x0000000000000000      0t0  15073670 type=STREAM
haproxy  32 root  124u     unix 0x0000000000000000      0t0  15071050 type=STREAM
haproxy  32 root  125u     IPv4           15073567      0t0       TCP 192.168.7.201:443->192.168.7.203:39386 (ESTABLISHED)
haproxy  32 root  126u     IPv4           15071791      0t0       TCP 192.168.7.201:443->192.168.7.203:39508 (ESTABLISHED)
haproxy  32 root  127u     unix 0x0000000000000000      0t0  15070935 type=STREAM
haproxy  32 root  128u     unix 0x0000000000000000      0t0  15073569 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  129u     unix 0x0000000000000000      0t0  15072693 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  130u     unix 0x0000000000000000      0t0  15071819 type=STREAM
haproxy  32 root  131u     IPv4           15071649      0t0       TCP 192.168.7.201:443->192.168.7.203:39402 (ESTABLISHED)
haproxy  32 root  132u     unix 0x0000000000000000      0t0  15073597 type=STREAM
haproxy  32 root  133u     unix 0x0000000000000000      0t0  15071651 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  134u     unix 0x0000000000000000      0t0  15070960 type=STREAM
haproxy  32 root  135u     unix 0x0000000000000000      0t0  15072701 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  136u     IPv4           15071661      0t0       TCP 192.168.7.201:443->192.168.7.203:39418 (ESTABLISHED)
haproxy  32 root  137u     unix 0x0000000000000000      0t0  15073602 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  138u     unix 0x0000000000000000      0t0  15071682 type=STREAM
haproxy  32 root  139u     IPv4           15073614      0t0       TCP 192.168.7.201:443->192.168.7.203:39444 (ESTABLISHED)
haproxy  32 root  140u     unix 0x0000000000000000      0t0  15070967 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  141u     unix 0x0000000000000000      0t0  15072746 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  142u     IPv4           15072616      0t0       TCP 192.168.7.201:443->192.168.7.203:39450 (ESTABLISHED)
haproxy  32 root  143u     unix 0x0000000000000000      0t0  15073631 type=STREAM
haproxy  32 root  144u     unix 0x0000000000000000      0t0  15074015 type=STREAM
haproxy  32 root  145u     IPv4           15072744      0t0       TCP 192.168.7.201:443->192.168.7.203:39534 (ESTABLISHED)
haproxy  32 root  146u     IPv4           15070987      0t0       TCP 192.168.7.201:443->192.168.7.203:39476 (ESTABLISHED)
haproxy  32 root  147u     unix 0x0000000000000000      0t0  15073649 type=STREAM
haproxy  32 root  148u     unix 0x0000000000000000      0t0  15072627 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  149u     unix 0x0000000000000000      0t0  15072641 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  150u     IPv4           15072011      0t0       TCP 192.168.7.201:443->192.168.7.203:39652 (ESTABLISHED)
haproxy  32 root  151u     unix 0x0000000000000000      0t0  15073010 type=STREAM
haproxy  32 root  152u     unix 0x0000000000000000      0t0  15071834 type=STREAM
haproxy  32 root  153u     unix 0x0000000000000000      0t0  15074096 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  154u     unix 0x0000000000000000      0t0  15072013 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  155u     IPv4           15071832      0t0       TCP 192.168.7.201:443->192.168.7.203:39536 (ESTABLISHED)
haproxy  32 root  156u     IPv4           15073015      0t0       TCP 192.168.7.201:443->192.168.7.203:39662 (ESTABLISHED)
haproxy  32 root  157u     unix 0x0000000000000000      0t0  15072056 type=STREAM
haproxy  32 root  158u     unix 0x0000000000000000      0t0  15072753 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  159u     IPv4           15072996      0t0       TCP 192.168.7.201:443->192.168.7.203:39634 (ESTABLISHED)
haproxy  32 root  160u     unix 0x0000000000000000      0t0  15073137 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  161u     IPv4           15074271      0t0       TCP 10.129.0.1:40710->10.130.0.63:8080 (ESTABLISHED)
haproxy  32 root  162u     unix 0x0000000000000000      0t0  15071864 type=STREAM
haproxy  32 root  163u     unix 0x0000000000000000      0t0  15072822 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  164u     IPv4           15073799      0t0       TCP 192.168.7.201:443->192.168.7.203:39562 (ESTABLISHED)
haproxy  32 root  165u     unix 0x0000000000000000      0t0  15074124 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  166u     unix 0x0000000000000000      0t0  15072998 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  167u     unix 0x0000000000000000      0t0  15073816 type=STREAM
haproxy  32 root  168u     IPv4           15073814      0t0       TCP 192.168.7.201:443->192.168.7.203:39570 (ESTABLISHED)
haproxy  32 root  169u     IPv4           15074122      0t0       TCP 192.168.7.201:443->192.168.7.203:39680 (ESTABLISHED)
haproxy  32 root  170u     unix 0x0000000000000000      0t0  15072836 type=STREAM
haproxy  32 root  171u     IPv4           15072834      0t0       TCP 192.168.7.201:443->192.168.7.203:39592 (ESTABLISHED)
haproxy  32 root  172u     unix 0x0000000000000000      0t0  15073830 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  173u     IPv4           15072966      0t0       TCP 192.168.7.201:443->192.168.7.203:39602 (ESTABLISHED)
haproxy  32 root  174u     unix 0x0000000000000000      0t0  15072968 type=STREAM
haproxy  32 root  175u     unix 0x0000000000000000      0t0  15071940 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  176u     IPv4           15074281      0t0       TCP 10.129.0.1:37816->10.128.0.60:8080 (ESTABLISHED)
haproxy  32 root  177u     IPv4           15073986      0t0       TCP 192.168.7.201:443->192.168.7.203:39624 (ESTABLISHED)
haproxy  32 root  178u     unix 0x0000000000000000      0t0  15071954 type=STREAM
haproxy  32 root  179u     IPv4           15072155      0t0       TCP 192.168.7.201:443->192.168.7.203:39766 (ESTABLISHED)
haproxy  32 root  180u     unix 0x0000000000000000      0t0  15074200 type=STREAM
haproxy  32 root  181u     IPv4           15074138      0t0       TCP 192.168.7.201:443->192.168.7.203:39690 (ESTABLISHED)
haproxy  32 root  182u     unix 0x0000000000000000      0t0  15072074 type=STREAM
haproxy  32 root  183u     unix 0x0000000000000000      0t0  15072832 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  184u     unix 0x0000000000000000      0t0  15073064 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  185u     unix 0x0000000000000000      0t0  15075378 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  186u     IPv4           15074147      0t0       TCP 192.168.7.201:443->192.168.7.203:39708 (ESTABLISHED)
haproxy  32 root  187u     unix 0x0000000000000000      0t0  15074149 type=STREAM
haproxy  32 root  188u     unix 0x0000000000000000      0t0  15073070 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  189u     IPv4           15076384      0t0       TCP 10.129.0.1:37860->10.128.0.60:8080 (ESTABLISHED)
haproxy  32 root  190u     unix 0x0000000000000000      0t0  15074159 type=STREAM
haproxy  32 root  191u     unix 0x0000000000000000      0t0  15072618 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  192u     IPv4           15072101      0t0       TCP 192.168.7.201:443->192.168.7.203:39718 (ESTABLISHED)
haproxy  32 root  193u     unix 0x0000000000000000      0t0  15072103 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  194u     unix 0x0000000000000000      0t0  15072157 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  195u     unix 0x0000000000000000      0t0  15073254 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  196u     unix 0x0000000000000000      0t0  15073084 type=STREAM
haproxy  32 root  197u     IPv4           15074173      0t0       TCP 192.168.7.201:443->192.168.7.203:39736 (ESTABLISHED)
haproxy  32 root  198u     unix 0x0000000000000000      0t0  15072117 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  199u     IPv4           15076401      0t0       TCP 10.129.0.1:44742->10.129.0.195:8080 (ESTABLISHED)
haproxy  32 root  200u     IPv4           15073091      0t0       TCP 192.168.7.201:443->192.168.7.203:39748 (ESTABLISHED)
haproxy  32 root  201u     unix 0x0000000000000000      0t0  15074188 type=STREAM
haproxy  32 root  202u     unix 0x0000000000000000      0t0  15073093 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  203u     unix 0x0000000000000000      0t0  15069999 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  204u     IPv4           15073111      0t0       TCP 10.129.0.1:44656->10.129.0.195:8080 (ESTABLISHED)
haproxy  32 root  205u     unix 0x0000000000000000      0t0  15075453 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  206u     unix 0x0000000000000000      0t0  15074756 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root  208u     IPv4           15065443      0t0       TCP 10.129.0.1:47706->10.130.0.18:6443 (ESTABLISHED)
haproxy  32 root  209u     IPv4           15064776      0t0       TCP 192.168.7.201:443->192.168.7.203:38356 (ESTABLISHED)
haproxy  32 root  210u     unix 0x0000000000000000      0t0  15073988 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM

$ oc -n openshift-ingress rsh router-default-bcd788f9-2p5zj lsof -n -P -p 32 | grep ESTAB | wc -l
Defaulted container "router" out of: router, logs
69

We do accumulate open connections when DisableKeepAlives: false.

Looking at the established connections in each canary pod:

$ oc get pods -n openshift-ingress-canary
NAME                   READY   STATUS    RESTARTS   AGE
ingress-canary-2gzvd   1/1     Running   0          171m
ingress-canary-5wp9s   1/1     Running   0          171m
ingress-canary-wtp42   1/1     Running   0          171m

$ oc rsh -n openshift-ingress-canary ingress-canary-2gzvd lsof -p 1 -n -P
COMMAND   PID       USER   FD      TYPE   DEVICE SIZE/OFF      NODE NAME
ingress-o   1 1000620000  cwd       DIR    0,440       17 226498931 /
ingress-o   1 1000620000  rtd       DIR    0,440       17 226498931 /
ingress-o   1 1000620000  txt       REG    0,440 79323386  44076225 /usr/bin/ingress-operator
ingress-o   1 1000620000  mem       REG    252,4           44076225 /usr/bin/ingress-operator (path dev=0,440)
ingress-o   1 1000620000    0u      CHR      1,3      0t0  13204983 /dev/null
ingress-o   1 1000620000    1w     FIFO     0,13      0t0  13204915 pipe
ingress-o   1 1000620000    2w     FIFO     0,13      0t0  13204916 pipe
ingress-o   1 1000620000    3u     IPv6 13205028      0t0       TCP *:8888 (LISTEN)
ingress-o   1 1000620000    4u  a_inode     0,14        0     11328 [eventpoll]
ingress-o   1 1000620000    5r     FIFO     0,13      0t0  13205024 pipe
ingress-o   1 1000620000    6w     FIFO     0,13      0t0  13205024 pipe
ingress-o   1 1000620000    7u     IPv6 13205033      0t0       TCP *:8080 (LISTEN)
ingress-o   1 1000620000    8u     IPv6 15121783      0t0       TCP 10.129.0.195:8080->10.129.0.1:51876 (ESTABLISHED)

$ oc rsh -n openshift-ingress-canary ingress-canary-5wp9s lsof -p 1 -n -P
COMMAND   PID       USER   FD      TYPE   DEVICE SIZE/OFF     NODE NAME
ingress-o   1 1000620000  cwd       DIR   0,1080       17 92277957 /
ingress-o   1 1000620000  rtd       DIR   0,1080       17 92277957 /
ingress-o   1 1000620000  txt       REG   0,1080 79323386 12585109 /usr/bin/ingress-operator
ingress-o   1 1000620000  mem       REG    252,4          12585109 /usr/bin/ingress-operator (path dev=0,1080)
ingress-o   1 1000620000    0u      CHR      1,3      0t0  9812059 /dev/null
ingress-o   1 1000620000    1w     FIFO     0,13      0t0  9807856 pipe
ingress-o   1 1000620000    2w     FIFO     0,13      0t0  9807857 pipe
ingress-o   1 1000620000    3u     IPv6  9812997      0t0      TCP *:8080 (LISTEN)
ingress-o   1 1000620000    4u  a_inode     0,14        0    10334 [eventpoll]
ingress-o   1 1000620000    5r     FIFO     0,13      0t0  9810897 pipe
ingress-o   1 1000620000    6w     FIFO     0,13      0t0  9810897 pipe
ingress-o   1 1000620000    7u     IPv6 11260627      0t0      TCP 10.128.0.60:8080->10.129.0.1:46206 (ESTABLISHED)
ingress-o   1 1000620000    8u     IPv6  9810901      0t0      TCP *:8888 (LISTEN)
ingress-o   1 1000620000    9u     IPv6 11261497      0t0      TCP 10.128.0.60:8080->10.129.0.1:46246 (ESTABLISHED)

$ oc rsh -n openshift-ingress-canary ingress-canary-wtp42 lsof -p 1 -n -P
COMMAND   PID       USER   FD      TYPE   DEVICE SIZE/OFF      NODE NAME
ingress-o   1 1000620000  cwd       DIR    0,483       17 224506323 /
ingress-o   1 1000620000  rtd       DIR    0,483       17 224506323 /
ingress-o   1 1000620000  txt       REG    0,483 79323386   8401671 /usr/bin/ingress-operator
ingress-o   1 1000620000  mem       REG    252,4            8401671 /usr/bin/ingress-operator (path dev=0,483)
ingress-o   1 1000620000    0u      CHR      1,3      0t0  10536382 /dev/null
ingress-o   1 1000620000    1w     FIFO     0,13      0t0  10537262 pipe
ingress-o   1 1000620000    2w     FIFO     0,13      0t0  10537263 pipe
ingress-o   1 1000620000    3u     IPv6 10536433      0t0       TCP *:8888 (LISTEN)
ingress-o   1 1000620000    4u  a_inode     0,14        0     10334 [eventpoll]
ingress-o   1 1000620000    5r     FIFO     0,13      0t0  10536428 pipe
ingress-o   1 1000620000    6w     FIFO     0,13      0t0  10536428 pipe
ingress-o   1 1000620000    7u     IPv6 10539336      0t0       TCP *:8080 (LISTEN)
ingress-o   1 1000620000    8u     IPv6 12236412      0t0       TCP 10.130.0.63:8080->10.129.0.1:49868 (ESTABLISHED)

I don't see any evidence that connections are held open by the canary
healthcheck.

Looking at the open connections the ingress-operator has I see:

$ ps -ef| grep ingress-operator
aim       160192  160090  0 12:04 pts/0    00:00:03 ./ingress-operator start --image quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:731bbf76d199bd256d91fe86c3be422bdba687af4f46ee5bfafde0684ac7d0c4 --canary-image=quay.io/openshift-release-dev/ocp-v4.0-art-dev@sha256:d8d30eb5dbd8a073ba3fa28a000795a474e0dd114a5f81bef2c1742b08a93e03 --release-version 4.11.0-0.nightly-2022-03-16-103946 --namespace openshift-ingress-operator --shutdown-file

$ lsof -p 160192

COMMAND      PID USER   FD      TYPE  DEVICE SIZE/OFF     NODE NAME
ingress-o 160192  aim  cwd       DIR    0,44      444  3884792 /home/aim/src/github.com/openshift/cluster-ingress-operator
ingress-o 160192  aim  rtd       DIR    0,35      168      256 /
ingress-o 160192  aim  txt       REG    0,44 79428664 26471552 /home/aim/src/github.com/openshift/cluster-ingress-operator/ingress-operator
ingress-o 160192  aim  mem       REG    0,42          26471552 /home/aim/src/github.com/openshift/cluster-ingress-operator/ingress-operator (path dev=0,44)
ingress-o 160192  aim    0u      CHR   136,0      0t0        3 /dev/pts/0
ingress-o 160192  aim    1u      CHR   136,0      0t0        3 /dev/pts/0
ingress-o 160192  aim    2u      CHR   136,0      0t0        3 /dev/pts/0
ingress-o 160192  aim    3u     IPv4 1646848      0t0      TCP route2.int.frobware.com:34416->lb-ocp411.ocp411.int.frobware.com:sun-sr-https (ESTABLISHED)
ingress-o 160192  aim    4u  a_inode    0,14        0    12022 [eventpoll:5,12,20,21,22,24,25,27,32,33,34,35,37,39,43,44,45,46,47,48,51,53,55,56,57,58,59,60,63,64,65,68...]
ingress-o 160192  aim    5r     FIFO    0,13      0t0  1664618 pipe
ingress-o 160192  aim    6w     FIFO    0,13      0t0  1664618 pipe
ingress-o 160192  aim    7r  a_inode    0,14        0    12022 inotify
ingress-o 160192  aim    8u  a_inode    0,14        0    12022 [eventpoll:7,9]
ingress-o 160192  aim    9r     FIFO    0,13      0t0  1634289 pipe
ingress-o 160192  aim   10w     FIFO    0,13      0t0  1634289 pipe
ingress-o 160192  aim   11u     IPv4 1636161      0t0      TCP localhost:60000 (LISTEN)
ingress-o 160192  aim   12u     IPv6 1648969      0t0      TCP *:webcache (LISTEN)
ingress-o 160192  aim   13u     IPv4 1669483      0t0      TCP route2.int.frobware.com:50384->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   14u     IPv4 1666506      0t0      TCP route2.int.frobware.com:50386->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   15u     IPv4 1653182      0t0      TCP route2.int.frobware.com:50388->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   16u     IPv4 1647357      0t0      TCP route2.int.frobware.com:50390->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   17u     IPv4 1647358      0t0      TCP route2.int.frobware.com:50392->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   18u     IPv4 1669494      0t0      TCP route2.int.frobware.com:50396->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   19u     IPv4 1671453      0t0      TCP route2.int.frobware.com:50398->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   20u     IPv4 1672437      0t0      TCP route2.int.frobware.com:50400->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   21u     IPv4 1670465      0t0      TCP route2.int.frobware.com:50402->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   22u     IPv4 1670466      0t0      TCP route2.int.frobware.com:50404->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   23u     IPv4 1650435      0t0      TCP route2.int.frobware.com:50406->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   24u     IPv4 1658786      0t0      TCP route2.int.frobware.com:50408->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   25u     IPv4 1672438      0t0      TCP route2.int.frobware.com:50410->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   26u     IPv4 1669497      0t0      TCP route2.int.frobware.com:50414->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   27u     IPv4 1672439      0t0      TCP route2.int.frobware.com:50416->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   28u     IPv4 1652282      0t0      TCP route2.int.frobware.com:50418->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   29u     IPv4 1650437      0t0      TCP route2.int.frobware.com:50420->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   30u     IPv4 1647371      0t0      TCP route2.int.frobware.com:50422->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   31u     IPv4 1673460      0t0      TCP route2.int.frobware.com:50424->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   32u     IPv4 1672440      0t0      TCP route2.int.frobware.com:50426->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   33u     IPv4 1673462      0t0      TCP route2.int.frobware.com:50428->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   34u     IPv4 1673463      0t0      TCP route2.int.frobware.com:50430->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   35u     IPv4 1670468      0t0      TCP route2.int.frobware.com:50432->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   36u     IPv4 1652286      0t0      TCP route2.int.frobware.com:50434->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   37u     IPv4 1672444      0t0      TCP route2.int.frobware.com:50436->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   38u     IPv4 1671473      0t0      TCP route2.int.frobware.com:50438->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   39u     IPv4 1670469      0t0      TCP route2.int.frobware.com:50440->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   40u     IPv4 1647374      0t0      TCP route2.int.frobware.com:50444->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   41u     IPv4 1652288      0t0      TCP route2.int.frobware.com:50446->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   42u     IPv4 1667597      0t0      TCP route2.int.frobware.com:50448->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   43u     IPv4 1670471      0t0      TCP route2.int.frobware.com:50450->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   44u     IPv4 1652291      0t0      TCP route2.int.frobware.com:50454->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   45u     IPv4 1670473      0t0      TCP route2.int.frobware.com:50456->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   46u     IPv4 1666528      0t0      TCP route2.int.frobware.com:50458->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   47u     IPv4 1673475      0t0      TCP route2.int.frobware.com:50460->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   48u     IPv4 1672453      0t0      TCP route2.int.frobware.com:50464->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   49u     IPv4 1653197      0t0      TCP route2.int.frobware.com:50466->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   50u     IPv4 1669511      0t0      TCP route2.int.frobware.com:50468->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   51u     IPv4 1672456      0t0      TCP route2.int.frobware.com:50470->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   52u     IPv4 1650442      0t0      TCP route2.int.frobware.com:50472->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   53u     IPv4 1673482      0t0      TCP route2.int.frobware.com:50474->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   54u     IPv4 1654507      0t0      TCP route2.int.frobware.com:50476->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   55u     IPv4 1669513      0t0      TCP route2.int.frobware.com:50478->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   56u     IPv4 1669514      0t0      TCP route2.int.frobware.com:50482->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   57u     IPv4 1672457      0t0      TCP route2.int.frobware.com:50484->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   58u     IPv4 1665691      0t0      TCP route2.int.frobware.com:50486->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   59u     IPv4 1665692      0t0      TCP route2.int.frobware.com:50488->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   60u     IPv4 1669517      0t0      TCP route2.int.frobware.com:50490->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   61u     IPv4 1665700      0t0      TCP route2.int.frobware.com:50492->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   62u     IPv4 1675268      0t0      TCP route2.int.frobware.com:50494->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   63u     IPv4 1666534      0t0      TCP route2.int.frobware.com:50496->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   64u     IPv4 1673487      0t0      TCP route2.int.frobware.com:50498->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   65u     IPv4 1673489      0t0      TCP route2.int.frobware.com:50500->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   66u     IPv4 1667619      0t0      TCP route2.int.frobware.com:50502->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   67u     IPv4 1671493      0t0      TCP route2.int.frobware.com:50504->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   68u     IPv4 1650452      0t0      TCP route2.int.frobware.com:50506->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   69u     IPv4 1667623      0t0      TCP route2.int.frobware.com:50508->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)
ingress-o 160192  aim   70u     IPv4 1654515      0t0      TCP route2.int.frobware.com:50510->lb-ocp411.ocp411.int.frobware.com:https (ESTABLISHED)

$ lsof -p 160192 | grep ESTABLISHED | wc -l
59

Let's go back and rebuild the ingress-operator but with keepalives
disabled:

modified   pkg/operator/controller/canary/controller.go
@@ -36,7 +36,7 @@ import (
 const (
 	canaryControllerName = "canary_controller"
 	// canaryCheckFrequency is how long to wait in between canary checks.
-	canaryCheckFrequency = 1 * time.Minute
+	canaryCheckFrequency = 500 * time.Millisecond
 	// canaryCheckCycleCount is how many successful canary checks should be observed
 	// before rotating the canary endpoint.
 	canaryCheckCycleCount = 5

$ make
hack/update-generated-crd.sh
hack/update-profile-manifests.sh
hack/update-generated-bindata.sh
CGO_ENABLED=0 GO111MODULE=on GOFLAGS=-mod=vendor go build -o ingress-operator  github.com/openshift/cluster-ingress-operator/cmd/ingress-operator

Repeating some of what we've already done. Let's look at haproxy again
to see how many open connections it has:

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj ps -ef
Defaulted container "router" out of: router, logs
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 11:50 ?        00:00:00 /usr/bin/openshift-router --v=2
root          32       1  0 11:50 ?        00:00:06 /usr/sbin/haproxy -f /var/lib/haproxy/conf/haproxy.config -p /var/lib/haproxy/run/haproxy.pid -x /var/lib/haproxy/run/haproxy.sock -sf 21
root          39       0  0 11:51 pts/0    00:00:00 /bin/sh
root         163       0  0 12:14 pts/1    00:00:00 ps -ef

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj lsof -p 32 -n -P
COMMAND PID USER   FD      TYPE             DEVICE SIZE/OFF      NODE NAME
haproxy  32 root  cwd       DIR              0,601     4096 127954449 /var/lib/haproxy/conf
haproxy  32 root  rtd       DIR              0,601       62 201331067 /
haproxy  32 root  txt       REG              0,601  2546664   8400718 /usr/sbin/haproxy
haproxy  32 root  mem       REG              252,4            8400718 /usr/sbin/haproxy (path dev=0,601)
haproxy  32 root  DEL       REG                0,1           14883247 /dev/zero
haproxy  32 root  mem       REG              252,4          127926420 /usr/lib64/libgcc_s-8-20200928.so.1 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127927128 /usr/lib64/libc-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951809 /usr/lib64/libpcre.so.1.2.10 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127927141 /usr/lib64/libpcreposix.so.0.0.6 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127952570 /usr/lib64/libcrypto.so.1.1.1g (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951466 /usr/lib64/libssl.so.1.1.1g (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951845 /usr/lib64/libpthread-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928166 /usr/lib64/librt-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127951435 /usr/lib64/libdl-2.28.so (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928156 /usr/lib64/libz.so.1.2.11 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127928149 /usr/lib64/libcrypt.so.1.1.0 (path dev=0,601)
haproxy  32 root  mem       REG              252,4          127926451 /usr/lib64/ld-2.28.so (path dev=0,601)
haproxy  32 root    0u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    1u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    2u      CHR                1,3      0t0  14879468 /dev/null
haproxy  32 root    3u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root    4u     unix 0x0000000000000000      0t0  14883251 type=DGRAM
haproxy  32 root    5u     unix 0x0000000000000000      0t0  14879703 /var/lib/haproxy/run/haproxy.sock.19.tmp type=STREAM
haproxy  32 root    6u     IPv4           14879706      0t0       TCP *:80 (LISTEN)
haproxy  32 root    7u     IPv4           14879708      0t0       TCP *:443 (LISTEN)
haproxy  32 root    8u     unix 0x0000000000000000      0t0  14879709 /var/lib/haproxy/run/haproxy-sni.sock.19.tmp type=STREAM
haproxy  32 root    9u     unix 0x0000000000000000      0t0  14879710 /var/lib/haproxy/run/haproxy-no-sni.sock.19.tmp type=STREAM
haproxy  32 root   10r     FIFO               0,13      0t0  14885925 pipe
haproxy  32 root   11w     FIFO               0,13      0t0  14885925 pipe
haproxy  32 root   12u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   13r     FIFO               0,13      0t0  14884953 pipe
haproxy  32 root   14w     FIFO               0,13      0t0  14884953 pipe
haproxy  32 root   15r     FIFO               0,13      0t0  14885926 pipe
haproxy  32 root   16w     FIFO               0,13      0t0  14885926 pipe
haproxy  32 root   17u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   18r     FIFO               0,13      0t0  14884107 pipe
haproxy  32 root   19w     FIFO               0,13      0t0  14884107 pipe
haproxy  32 root   20u  a_inode               0,14        0     11328 [eventpoll]
haproxy  32 root   22u     unix 0x0000000000000000      0t0  14885999 type=DGRAM
haproxy  32 root   23u     unix 0x0000000000000000      0t0  14884315 type=DGRAM
haproxy  32 root   24u     unix 0x0000000000000000      0t0  14886722 type=DGRAM
haproxy  32 root   27u     IPv4           15189426      0t0       TCP 10.129.0.1:58052->10.130.0.63:8080 (ESTABLISHED)
haproxy  32 root   29u     IPv4           15189229      0t0       TCP 10.129.0.1:55006->10.128.0.60:8080 (ESTABLISHED)
haproxy  32 root   30u     IPv4           15098095      0t0       TCP 10.129.0.1:53074->10.130.0.18:6443 (ESTABLISHED)
haproxy  32 root   31u     IPv4           15098093      0t0       TCP 192.168.7.201:443->192.168.7.203:41804 (ESTABLISHED)
haproxy  32 root   32u     IPv4           15191374      0t0       TCP 10.129.0.1:33770->10.129.0.195:8080 (ESTABLISHED)
haproxy  32 root   34u     IPv4           15191196      0t0       TCP 10.129.0.1:55038->10.128.0.60:8080 (ESTABLISHED)
haproxy  32 root   37u     IPv4           15192303      0t0       TCP 10.129.0.1:55108->10.128.0.60:8080 (ESTABLISHED)
haproxy  32 root  207u     IPv4           15153094      0t0       TCP 192.168.7.201:443->192.168.7.203:49520 (ESTABLISHED)
haproxy  32 root  208u     IPv4           15154925      0t0       TCP 10.129.0.1:33036->10.130.0.18:6443 (ESTABLISHED)

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj lsof -p 32 -n -P | grep ESTAB | wc -l
Defaulted container "router" out of: router, logs
14

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj lsof -p 32 -n -P | grep ESTAB | wc -l
Defaulted container "router" out of: router, logs
9

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj lsof -p 32 -n -P | grep ESTAB | wc -l
Defaulted container "router" out of: router, logs
8

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj lsof -p 32 -n -P | grep ESTAB | wc -l
Defaulted container "router" out of: router, logs
10

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj lsof -p 32 -n -P | grep ESTAB | wc -l
Defaulted container "router" out of: router, logs
13

$ oc rsh -n openshift-ingress router-default-bcd788f9-2p5zj lsof -p 32 -n -P | grep ESTAB | wc -l
Defaulted container "router" out of: router, logs
7

Significantly fewer. At the moment it's not clear to me that this is a
bug. With "DisableKeepAlives: true" we have significantly fewer open
and established connections.

I will follow up with some additional debugging that focuses on packet
captures. 

