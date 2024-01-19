Futzing with certficates & routes & services

# Install

	$ oc apply -f ne1444-service.yaml
	$ ./process-routes.sh

Verify reencrypt works OK

	$ curl -k -vv https://ne1444-default-cert-reencrypt-default.apps.ocp.frobware.lan
	*   Trying 192.168.7.164:443...
	* Connected to ne1444-default-cert-reencrypt-default.apps.ocp.frobware.lan (192.168.7.164) port 443 (#0)
	* ALPN, offering h2
	* ALPN, offering http/1.1
	* successfully set certificate verify locations:
	*  CAfile: /etc/pki/tls/certs/ca-bundle.crt
	*  CApath: none
	* TLSv1.3 (OUT), TLS handshake, Client hello (1):
	* TLSv1.3 (IN), TLS handshake, Server hello (2):
	* TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
	* TLSv1.3 (IN), TLS handshake, Certificate (11):
	* TLSv1.3 (IN), TLS handshake, CERT verify (15):
	* TLSv1.3 (IN), TLS handshake, Finished (20):
	* TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
	* TLSv1.3 (OUT), TLS handshake, Finished (20):
	* SSL connection using TLSv1.3 / TLS_AES_256_GCM_SHA384
	* ALPN, server did not agree to a protocol
	* Server certificate:
	*  subject: O=Cert Gen Company; CN=Cert Gen Company Common Name
	*  start date: Jul 29 14:16:14 2021 GMT
	*  expire date: Jul  5 14:16:14 2121 GMT
	*  issuer: O=Cert Gen Co; CN=Root CA
	*  SSL certificate verify result: self signed certificate in certificate chain (19), continuing anyway.
	> GET / HTTP/1.1
	> Host: ne1444-default-cert-reencrypt-default.apps.ocp.frobware.lan
	> User-Agent: curl/7.76.1
	> Accept: */*
	>
	* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
	* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
	* old SSL session ID is stale, removing
	* Mark bundle as not supporting multiuse
	< HTTP/1.1 200 OK
	< date: Thu, 29 Jul 2021 14:17:30 GMT
	< content-length: 8
	< content-type: text/plain; charset=utf-8
	< set-cookie: d45c8a9bc4613523546f27500f50681b=1223926b9585c1024e600b183e54f17d; path=/; HttpOnly; Secure; SameSite=None
	< cache-control: private
	<
	* Connection #0 to host ne1444-default-cert-reencrypt-default.apps.ocp.frobware.lan left intact

And verification from the service/pod:

	$ oc get pods -o wide
	NAME     READY   STATUS    RESTARTS   AGE   IP           NODE                        NOMINATED NODE   READINESS GATES
	ne1444   1/1     Running   0          94m   10.128.2.9   worker-1.ocp.frobware.lan   <none>           <none>

	$ oc rsh ne1444 
	sh-4.4# curl -k -vv https://10.128.2.9:8443
	* Rebuilt URL to: https://10.128.2.9:8443/
	*   Trying 10.128.2.9...
	* TCP_NODELAY set
	* Connected to 10.128.2.9 (10.128.2.9) port 8443 (#0)
	* ALPN, offering h2
	* ALPN, offering http/1.1
	* successfully set certificate verify locations:
	*   CAfile: /etc/pki/tls/certs/ca-bundle.crt
	  CApath: none
	* TLSv1.3 (OUT), TLS handshake, Client hello (1):
	* TLSv1.3 (IN), TLS handshake, Server hello (2):
	* TLSv1.3 (IN), TLS handshake, [no content] (0):
	* TLSv1.3 (IN), TLS handshake, Encrypted Extensions (8):
	* TLSv1.3 (IN), TLS handshake, [no content] (0):
	* TLSv1.3 (IN), TLS handshake, Certificate (11):
	* TLSv1.3 (IN), TLS handshake, [no content] (0):
	* TLSv1.3 (IN), TLS handshake, CERT verify (15):
	* TLSv1.3 (IN), TLS handshake, [no content] (0):
	* TLSv1.3 (IN), TLS handshake, Finished (20):
	* TLSv1.3 (OUT), TLS change cipher, Change cipher spec (1):
	* TLSv1.3 (OUT), TLS handshake, [no content] (0):
	* TLSv1.3 (OUT), TLS handshake, Finished (20):
	* SSL connection using TLSv1.3 / TLS_CHACHA20_POLY1305_SHA256
	* ALPN, server accepted to use h2
	* Server certificate:
	*  subject: CN=ne1444.default.svc
	*  start date: Jul 29 12:45:07 2021 GMT
	*  expire date: Jul 29 12:45:08 2023 GMT
	*  issuer: CN=openshift-service-serving-signer@1626459196
	*  SSL certificate verify result: self signed certificate in certificate chain (19), continuing anyway.
	* Using HTTP2, server supports multi-use
	* Connection state changed (HTTP/2 confirmed)
	* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
	* TLSv1.3 (OUT), TLS app data, [no content] (0):
	* TLSv1.3 (OUT), TLS app data, [no content] (0):
	* TLSv1.3 (OUT), TLS app data, [no content] (0):
	* Using Stream ID: 1 (easy handle 0x56392396b6b0)
	* TLSv1.3 (OUT), TLS app data, [no content] (0):
	> GET / HTTP/2
	> Host: 10.128.2.9:8443
	> User-Agent: curl/7.61.1
	> Accept: */*
	> 
	* TLSv1.3 (IN), TLS handshake, [no content] (0):
	* TLSv1.3 (IN), TLS handshake, Newsession Ticket (4):
	* TLSv1.3 (IN), TLS app data, [no content] (0):
	* Connection state changed (MAX_CONCURRENT_STREAMS == 250)!
	* TLSv1.3 (OUT), TLS app data, [no content] (0):
	* TLSv1.3 (IN), TLS app data, [no content] (0):
	* TLSv1.3 (IN), TLS app data, [no content] (0):
	< HTTP/2 200 
	< content-type: text/plain; charset=utf-8
	< content-length: 8
	< date: Thu, 29 Jul 2021 14:19:40 GMT
	< 
	* Connection #0 to host 10.128.2.9 left intact
	HTTP/2.0sh-4.4# 
