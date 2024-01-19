// package main

// import (
// 	"crypto/tls"
// 	"crypto/x509"
// 	"fmt"
// 	"os"

// 	"github.com/google/go-cmp/cmp"
// )

// func getSSLCertificate(host string) (*x509.Certificate, error) {
// 	conn, err := tls.Dial("tcp", host+":443", &tls.Config{
// 		InsecureSkipVerify: true,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer conn.Close()

// 	certs := conn.ConnectionState().PeerCertificates
// 	if len(certs) == 0 {
// 		return nil, fmt.Errorf("no certificates found")
// 	}

// 	return certs[0], nil
// }

// func main() {
// 	if len(os.Args) != 3 {
// 		fmt.Println("Usage: go run script.go <host1> <host2>")
// 		os.Exit(1)
// 	}

// 	cert1, err := getSSLCertificate(os.Args[1])
// 	if err != nil {
// 		fmt.Printf("Error retrieving certificate from %s: %v\n", os.Args[1], err)
// 		os.Exit(1)
// 	}

// 	cert2, err := getSSLCertificate(os.Args[2])
// 	if err != nil {
// 		fmt.Printf("Error retrieving certificate from %s: %v\n", os.Args[2], err)
// 		os.Exit(1)
// 	}

// 	diff := cmp.Diff(cert1, cert2)
// 	if diff == "" {
// 		fmt.Println("No differences found.")
// 	} else {
// 		fmt.Println("Differences found:")
// 		fmt.Println(diff)
// 	}
// }

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
)

func getSSLCertificate(host string) (*x509.Certificate, error) {
	conn, err := tls.Dial("tcp", host+":443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	return certs[0], nil
}

func compareCertificates(cert1, cert2 *x509.Certificate) []string {
	var differences []string

	// Compare Issuer
	if cert1.Issuer.String() != cert2.Issuer.String() {
		differences = append(differences, fmt.Sprintf("Issuer: %s != %s", cert1.Issuer, cert2.Issuer))
	}

	// Compare Subject
	if cert1.Subject.String() != cert2.Subject.String() {
		differences = append(differences, fmt.Sprintf("Subject: %s != %s", cert1.Subject, cert2.Subject))
	}

	// Compare Validity Dates
	if !cert1.NotBefore.Equal(cert2.NotBefore) || !cert1.NotAfter.Equal(cert2.NotAfter) {
		differences = append(differences, fmt.Sprintf("Validity: %s - %s != %s - %s", cert1.NotBefore, cert1.NotAfter, cert2.NotBefore, cert2.NotAfter))
	}

	// Compare Common Names
	if cert1.Subject.CommonName != cert2.Subject.CommonName {
		differences = append(differences, fmt.Sprintf("Common Name: %s != %s", cert1.Subject.CommonName, cert2.Subject.CommonName))
	}

	// Compare SANs
	sans1 := strings.Join(cert1.DNSNames, ", ")
	sans2 := strings.Join(cert2.DNSNames, ", ")
	if sans1 != sans2 {
		differences = append(differences, fmt.Sprintf("SANs: %s != %s", sans1, sans2))
	}

	return differences
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run script.go <host1> <host2>")
		os.Exit(1)
	}

	cert1, err := getSSLCertificate(os.Args[1])
	if err != nil {
		fmt.Printf("Error retrieving certificate from %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}

	cert2, err := getSSLCertificate(os.Args[2])
	if err != nil {
		fmt.Printf("Error retrieving certificate from %s: %v\n", os.Args[2], err)
		os.Exit(1)
	}

	differences := compareCertificates(cert1, cert2)
	if len(differences) == 0 {
		fmt.Println("No differences found.")
	} else {
		fmt.Println("Differences found:")
		for _, diff := range differences {
			fmt.Println(diff)
		}
	}
}
