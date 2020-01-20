FROM centos:7

RUN yum install -y psmisc procps-ng rsyslog sysvinit-tools socat which lsof
RUN rpm -ivh https://github.com/frobware/haproxy-hacks/raw/master/RPMs/PROXMOX-haproxy20-2.0.12-2.el7.x86_64.rpm

RUN mkdir -p /var/lib/haproxy && \
    mkdir -p /var/lib/haproxy/run && \
    mkdir -p /var/lib/haproxy/router/{certs,cacerts,whitelists} && \
    mkdir -p /var/lib/haproxy/{conf/.tmp,run,bin,log} && \
    touch /var/lib/haproxy/conf/{{os_http_be,os_edge_reencrypt_be,os_tcp_be,os_sni_passthrough,os_route_http_redirect,cert_config,os_wildcard_domain}.map,haproxy.config}

#STOPSIGNAL SIGUSR2

COPY start-haproxy /
COPY reload-haproxy /
COPY haproxy.cfg /etc/haproxy/haproxy.cfg

COPY docker-entrypoint.sh /
# ENTRYPOINT ["/docker-entrypoint.sh"]
# CMD ["haproxy", "-f", "/etc/haproxy/haproxy.cfg"]

ENTRYPOINT ["/start-haproxy"]
#ENTRYPOINT ["/usr/sbin/haproxy", "-f", "/etc/haproxy/haproxy.cfg", "-d"]
