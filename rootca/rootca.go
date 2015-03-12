package rootca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"
)

type RootCA struct {
	ca       *x509.Certificate
	priv     *rsa.PrivateKey
	derBytes []byte
}

type certPem struct {
	certFile []byte
	keyFile  []byte
}

func NewCA(name string, vaildFor time.Duration, rsaBits int) (*RootCA, error) {
	template := x509.Certificate{
		IsCA:         true,
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{name},
		},
		NotBefore: time.Now().Add(-time.Duration(5 * time.Minute)),
		NotAfter:  time.Now().Add(vaildFor),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	ca, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, err
	}

	return &RootCA{ca, priv, derBytes}, nil
}

func (r *RootCA) Dump(filename string) error {
	outFile, err := os.Create(filename)
	defer outFile.Close()
	if err != nil {
		return err
	}
	pem.Encode(outFile, &pem.Block{Type: "CERTIFICATE", Bytes: r.derBytes})
	pem.Encode(outFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(r.priv)})
	return nil
}

func NewCAFromFile(filename string) (*RootCA, error) {
	var r RootCA
	var b *pem.Block
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	for {
		b, data = pem.Decode(data)
		if b == nil {
			break
		}
		if b.Type == "CERTIFICATE" {
			r.derBytes = b.Bytes
			ca, err := x509.ParseCertificate(r.derBytes)
			if err != nil {
				return nil, err
			}
			r.ca = ca
		} else if b.Type == "RSA PRIVATE KEY" {
			priv, err := x509.ParsePKCS1PrivateKey(b.Bytes)
			if err != nil {
				return nil, err
			}
			r.priv = priv
		}
	}
	return &r, nil
}

func getCommonName(domain string) (host string, err error) {
	eTLD_1, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return
	}

	prefix := strings.TrimRight(strings.TrimSuffix(domain, eTLD_1), ".")
	if strings.Contains(prefix, ".") {
		host = fmt.Sprintf("%s.%s", strings.SplitN(prefix, ".", 2)[1], eTLD_1)
	} else {
		host = eTLD_1
	}
	return
}

func (c *RootCA) issue(host string, vaildFor time.Duration, rsaBits int) (*certPem, error) {
	host, err := getCommonName(host)
	if err != nil {
		return nil, err
	}
	csrTemplate := &x509.CertificateRequest{
		Signature: []byte(host),
		Subject: pkix.Name{
			Country:      []string{"CN"},
			Organization: []string{host},
		},
		DNSNames:           []string{fmt.Sprintf("*.%s", host)},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		return nil, err
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, priv)
	if err != nil {
		return nil, err
	}

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, err
	}

	certTemplate := &x509.Certificate{
		Subject:            csr.Subject,
		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,
		SerialNumber:       big.NewInt(int64(time.Now().Nanosecond())),
		SignatureAlgorithm: x509.SHA256WithRSA,
		NotBefore:          time.Now().Add(-time.Duration(10 * time.Minute)).UTC(),
		NotAfter:           time.Now().Add(vaildFor),
		KeyUsage:           x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		DNSNames: []string{fmt.Sprintf("*.%s", host)},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, certTemplate, c.ca, csr.PublicKey, c.priv)
	if err != nil {
		return nil, err
	}

	certFile := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyFile := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return &certPem{certFile, keyFile}, nil
}

func (c *RootCA) Issue(host string, vaildFor time.Duration, rsaBits int) (*tls.Certificate, error) {
	pem, err := c.issue(host, vaildFor, rsaBits)
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair(pem.certFile, pem.keyFile)
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}

func (c *RootCA) IssueFile(host string, vaildFor time.Duration, rsaBits int) (string, error) {
	pem, err := c.issue(host, vaildFor, rsaBits)
	if err != nil {
		return "", err
	}

	commonname, err := getCommonName(host)
	if err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%s.crt", commonname)

	outFile, err := os.Create(filename)
	defer outFile.Close()
	if err != nil {
		return "", err
	}
	_, err = outFile.Write(pem.certFile)
	if err != nil {
		return "", err
	}
	_, err = outFile.Write(pem.keyFile)
	if err != nil {
		return "", err
	}
	return filename, nil
}
