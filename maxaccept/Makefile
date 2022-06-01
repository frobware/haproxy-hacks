analyze:
	./print-errors.sh responses.csv

.PHONY: http-scale-edge.json.js | /var/tmp/haproxy/run
http-scale-edge.json-with-hops.js:
	mb -d60 -i http-scale-edge.json.js -o responses.csv

.PHONY: http-scale-edge.json-no-hop.js | /var/tmp/haproxy/run
http-scale-edge.json-no-hop.js:
	mb -d60 -i $@ -o responses.csv

/var/tmp/haproxy/run:
	mkdir -p $@
