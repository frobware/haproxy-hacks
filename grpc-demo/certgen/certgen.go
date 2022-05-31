package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"strings"
	"time"
)

type Config struct {
	Organization          []string
	CommonName            string
	SubjectAlternateNames []string
	NotBefore             time.Time
	NotAfter              time.Time
}

// MarshalKeyToDERFormat converts the key to a string representation
// (SEC 1, ASN.1 DER form) suitable for dropping into a route's TLS
// key stanza.
func MarshalKeyToDERFormat(key *ecdsa.PrivateKey) (string, error) {
	data, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("unable to marshal private key: %v", err)
	}

	buf := &bytes.Buffer{}

	if err := pem.Encode(buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: data}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// MarshalCertToPEMString encodes derBytes to a PEM format suitable
// for dropping into a route's TLS certificate stanza.
func MarshalCertToPEMString(derBytes []byte) (string, error) {
	buf := &bytes.Buffer{}

	if err := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", fmt.Errorf("failed to encode cert data: %v", err)
	}

	return buf.String(), nil
}

// GenerateKeyPair creates CA-Cert, certificate and key.
func GenerateKeyPair(cfg Config) ([]byte, []byte, *ecdsa.PrivateKey, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	rootKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate ECDSA key: %v", err)
	}

	rootTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Cert Gen Co"},
			CommonName:   "Root CA",
		},
		NotBefore:             cfg.NotBefore,
		NotAfter:              cfg.NotAfter,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCert, err := x509.CreateCertificate(rand.Reader, &rootTemplate, &rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create root certificate: %v", err)
	}

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate ECDSA key: %v", err)
	}

	serialNumber, err = rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	leafCertTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: cfg.Organization,
			CommonName:   cfg.CommonName,
		},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             cfg.NotBefore,
		NotAfter:              cfg.NotAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	for _, h := range cfg.SubjectAlternateNames {
		if ip := net.ParseIP(h); ip != nil {
			leafCertTemplate.IPAddresses = append(leafCertTemplate.IPAddresses, ip)
		} else {
			leafCertTemplate.DNSNames = append(leafCertTemplate.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &leafCertTemplate, &rootTemplate, &leafKey.PublicKey, rootKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create leaf certificate: %v", err)
	}

	return caCert, derBytes, leafKey, nil
}

func main() {
	notBefore := time.Now()

	cfg := Config{
		Organization:          []string{"Cert Gen Company"},
		CommonName:            "Cert Gen Company Common Name",
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(100 * time.Hour * 24 * 365), // 100 years
		SubjectAlternateNames: flag.Args(),
	}

	caCrt, crt, key, err := GenerateKeyPair(cfg)

	if err != nil {
		log.Fatalf("failed to generate key pair: %v", err)
	}

	s1, err := MarshalKeyToDERFormat(key)
	if err != nil {
		log.Fatalf("failed to marshal key: %v", err)
	}

	s2, err := MarshalCertToPEMString(crt)
	if err != nil {
		log.Fatalf("failed to marshal crt: %v", err)
	}

	s3, err := MarshalCertToPEMString(caCrt)
	if err != nil {
		log.Fatalf("failed to marshal CA cert: %v", err)
	}

	s1 = strings.TrimRight(s1, "\n")
	s2 = strings.TrimRight(s2, "\n")
	s3 = strings.TrimRight(s3, "\n")

	fmt.Printf("TLS_KEY=\"%s\"\n", s1)
	fmt.Printf("TLS_CRT=\"%s\"\n", s2)
	fmt.Printf("TLS_CACRT=\"%s\"\n", s3)
}
