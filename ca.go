package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

type CA struct {
	name         string
	rsaBits      int
	rootDerBytes []byte
	rootCert     *x509.Certificate
	rootPriv     *rsa.PrivateKey
	certs        map[string]*tls.Certificate
}

func NewCA(name string, rsaBits int) *CA {
	ca := &CA{
		name:    name,
		rsaBits: rsaBits,
	}
	ca.certs = make(map[string]*tls.Certificate)
	return ca
}

// http://golang.org/src/pkg/crypto/tls/generate_cert.go
func (c *CA) Create(fileName string, vaildFor time.Duration) {
	template := x509.Certificate{
		IsCA:         true,
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{c.name},
		},
		NotBefore: time.Now().Add(-time.Duration(5 * time.Minute)),
		NotAfter:  time.Now().Add(vaildFor),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, c.rsaBits)
	if err != nil {
		log.Fatalf("failed to generate private key: %s", err)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	outFile, err := os.Create(fileName)
	defer outFile.Close()
	if err != nil {
		log.Fatalf("failed to open cert.crt for writing: %s", err)
	}
	pem.Encode(outFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	pem.Encode(outFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	rootCert, _ := x509.ParseCertificate(derBytes)
	c.rootCert = rootCert
	rootPriv, _ := x509.ParsePKCS1PrivateKey(derBytes)
	c.rootPriv = rootPriv
}

// https://github.com/coreos/etcd-ca/blob/master/pkix/cert_host.go
func (c *CA) Issue(host string, vaildFor time.Duration) (*tls.Certificate, error) {
	csrTemplate := &x509.CertificateRequest{
		Signature: []byte(host),
		Subject: pkix.Name{
			Country:      []string{"CN"},
			Organization: []string{host},
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
		DNSNames:           []string{host},
		EmailAddresses:     []string{host},
		IPAddresses:        []net.IP{},
	}

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(int64(time.Now().Nanosecond())),
		Subject:      pkix.Name{},
		NotBefore:    time.Now().Add(-time.Duration(10 * time.Minute)).UTC(),
		NotAfter:     time.Now().Add(vaildFor),
		KeyUsage:     0,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		UnknownExtKeyUsage:          nil,
		BasicConstraintsValid:       false,
		SubjectKeyId:                nil,
		DNSNames:                    nil,
		PermittedDNSDomainsCritical: false,
		PermittedDNSDomains:         nil,
	}

	priv, err := rsa.GenerateKey(rand.Reader, c.rsaBits)
	if err != nil {
		log.Fatalf("failed to GenerateKey: %s", err)
	}

	derBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, priv)
	if err != nil {
		log.Fatalf("failed to CreateCertificateRequest: %s", err)
	}

	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		log.Fatalf("failed to CreateCertificateRequest: %s", err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, c.rootCert, pub, c.rootPriv)
	if err != nil {
		log.Fatal("failed to CreateCertificate: %s", err)
	}

	certPEM := make([]byte, 10240)
	certPEMWriter := bufio.NewWriter(bytes.NewBuffer(certPEM))
	pem.Encode(certPEMWriter, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	pem.Encode(certPEMWriter, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	certPEMWriter.Flush()

	tlsCert, err := tls.X509KeyPair(certPEM, certPEM)
	if err != nil {
		log.Fatal("failed to CreateCertificate: %s", err)
	} else {
		c.certs[host] = &tlsCert
	}
	return &tlsCert, err
}
