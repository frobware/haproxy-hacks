build:
	podman build . -t amcdermo/bz2044682-server

push: build
	podman tag amcdermo/bz2044682-server quay.io/amcdermo/bz2044682-server
	podman push quay.io/amcdermo/bz2044682-server:latest

clean:
	$(RM) *.yaml
