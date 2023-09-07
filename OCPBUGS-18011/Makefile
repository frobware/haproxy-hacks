all: results.org results.html

results.org:
	cat ./javascript.js > $@
	./results-7-sept.pl >> $@
	emacs --batch --load eval-table.el

results.html: results.org
	emacs --batch --file $< --eval "(org-html-export-to-html)"

.PHONY: results.org
