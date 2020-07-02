build:
	docker build . -t sqlite3-histo-builder

run:
	docker run --rm -v /tmp:/target sqlite3-histo-builder
	cp /tmp/sqlite3 ../curler
