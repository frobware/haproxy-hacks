Futzing with certficates & routes & services

# Install

	$ oc apply -f destca-service.yaml
	$ ./process-routes.sh

Verify reencrypt works OK

	$ curl -k -vv https://destca-default-cert-reencrypt-default.apps.ocp.frobware.lan
	*   Trying 192.168.7.164:443...
	* Connected to destca-default-cert-reencrypt-default.apps.ocp.frobware.lan (192.168.7.164) port 443 (#0)
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
	> Host: destca-default-cert-reencrypt-default.apps.ocp.frobware.lan
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
	* Connection #0 to host destca-default-cert-reencrypt-default.apps.ocp.frobware.lan left intact
