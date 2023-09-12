all: results.org results.html

results.org:
	cat ./javascript.js > $@
	#./results-7-sept.pl >> $@
	#./results-8-sept-local-clusters.pl >> $@
	./aws-results-12-sept.pl >> $@
	emacs --batch --load eval-table.el

results.html: results.org
	emacs --batch --file $< --eval "(org-html-export-to-html)"

.PHONY: results.org
