FROM centos:7

RUN yum install -y procps-ng rsyslog sysvinit-tools
RUN rpm -ivh http://spicy.frobware.com/~aim/x86_64/haproxy21-2.1.2-1.el7.x86_64.rpm

RUN mkdir -p /var/lib/haproxy && \
    mkdir -p /var/lib/haproxy/run && \
    mkdir -p /var/lib/haproxy/router/{certs,cacerts,whitelists} && \
    mkdir -p /var/lib/haproxy/{conf/.tmp,run,bin,log} && \
    touch /var/lib/haproxy/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}

COPY reload-haproxy /bin/
COPY haproxy.cfg /etc/haproxy/haproxy.cfg
COPY reaper /bin

ENTRYPOINT ["/bin/reaper"]
