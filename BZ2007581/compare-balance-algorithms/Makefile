all: clean
	./run-benchmark-scenarios.bash
	$(MAKE) summary

.PHONY:
clean:
	rm -rf benchmark-config*

.PHONY:
summary:
	./make-summary.md
