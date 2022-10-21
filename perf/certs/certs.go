package certs

import (
	"path/filepath"
	"runtime"
)

var basepath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(currentFile)
}

func TLSKeyFile() string {
	return filepath.Join(basepath, "testdata", "tls.key")
}

func TLSCertFile() string {
	return filepath.Join(basepath, "testdata", "tls.crt")
}
