# Building

## Fedora

	$ release=34
	$ toolbox create --release=${release} --distro=fedora
	$ toolbox enter fedora-toolbox-${release}
	$ sudo dnf install -y openssl-devel pcre-devel zlib-devel
	$ sudo dnf groupinstall -y "C Development Tools and Libraries"

# Debug support

## Fedora 32

	$ sudo dnf debuginfo-install -y libgcc-10.3.1-1.fc32.x86_64 libxcrypt-4.4.20-2.fc32.x86_64 openssl-libs-1.1.1k-1.fc32.x86_64 pcre-8.44-2.fc32.x86_64 zlib-1.2.11-21.fc32.x86_64

## Fedora 34

	$ sudo dnf debuginfo-install -y libgcc-11.2.1-1.fc34.x86_64 libxcrypt-4.4.26-2.fc34.x86_64 openssl-libs-1.1.1l-2.fc34.x86_64 pcre-8.44-3.fc34.1.x86_64 zlib-1.2.11-26.fc34.x86_64

# ubi8

	$ sudo yum debuginfo-install -y libgcc-8.4.1-1.el8.x86_64 libxcrypt-4.1.1-4.el8.x86_64 openssl-libs-1.1.1g-15.el8_3.x86_64 pcre-8.42-4.el8.x86_64 zlib-1.2.11-17.el8.x86_64
