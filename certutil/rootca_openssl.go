package certutil

import (
	"crypto/tls"
	"fmt"
	"github.com/phuslu/openssl"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

type OpenCA struct {
	privKey openssl.PrivateKey
	cert    *openssl.Certificate
}

type openCert struct {
	privKey openssl.PrivateKey
	cert    *openssl.Certificate
}

func NewOpenCA(name string, vaildFor time.Duration, rsaBits int) (CA, error) {
	privKey, err := openssl.GenerateRSAKey(rsaBits)
	if err != nil {
		return nil, err
	}

	info := &openssl.CertificateInfo{
		Serial:       big.NewInt(int64(1)),
		Issued:       0,
		Expires:      3 * 365 * 24 * time.Hour,
		Country:      "CN",
		Organization: name,
		CommonName:   name,
	}
	cert, err := openssl.NewCertificate(info, privKey)
	if err != nil {
		return nil, err
	}
	err = cert.AddExtensions(map[openssl.NID]string{
		openssl.NID_basic_constraints:      "critical,CA:TRUE",
		openssl.NID_key_usage:              "critical,keyCertSign,cRLSign",
		openssl.NID_subject_key_identifier: "hash",
		openssl.NID_netscape_cert_type:     "sslCA"})
	if err != nil {
		return nil, err
	}

	err = cert.Sign(privKey, openssl.EVP_SHA256)
	if err != nil {
		return nil, err
	}

	return &OpenCA{
		privKey: privKey,
		cert:    cert,
	}, nil
}

func NewOpenCAFromFile(filename string) (CA, error) {
	pem_block, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cert, err := openssl.LoadCertificateFromPEM(pem_block)
	if err != nil {
		return nil, err
	}

	privKey, err := openssl.LoadPrivateKeyFromPEM(pem_block)
	if err != nil {
		return nil, err
	}

	return &OpenCA{
		privKey: privKey,
		cert:    cert,
	}, nil
}

func (ca *OpenCA) Dump(filename string) error {
	outFile, err := os.Create(filename)
	defer outFile.Close()
	if err != nil {
		return err
	}
	certBytes, err := ca.cert.MarshalPEM()
	if err != nil {
		return err
	}
	_, err = outFile.Write(certBytes)
	if err != nil {
		return err
	}
	privBytes, err := ca.privKey.MarshalPKCS1PrivateKeyPEM()
	if err != nil {
		return err
	}
	_, err = outFile.Write(privBytes)
	if err != nil {
		return err
	}
	return nil
}

func (ca *OpenCA) issue(host string, vaildFor time.Duration, rsaBits int) (*openCert, error) {
	host, err := getCommonName(host)
	if err != nil {
		return nil, err
	}

	privKey, err := openssl.GenerateRSAKey(rsaBits)
	if err != nil {
		return nil, err
	}

	info := &openssl.CertificateInfo{
		Serial:       big.NewInt(time.Now().UnixNano()),
		Issued:       0,
		Expires:      3 * 365 * 24 * time.Hour,
		Country:      "CN",
		Organization: host,
		CommonName:   host,
	}
	cert, err := openssl.NewCertificate(info, privKey)
	if err != nil {
		return nil, err
	}

	err = cert.AddExtensions(map[openssl.NID]string{
		openssl.NID_subject_alt_name:  fmt.Sprintf("DNS:*.%s", host),
		openssl.NID_basic_constraints: "critical,CA:FALSE",
		openssl.NID_key_usage:         "keyEncipherment",
		openssl.NID_ext_key_usage:     "serverAuth"})
	if err != nil {
		return nil, err
	}

	err = cert.SetIssuer(ca.cert)
	if err != nil {
		return nil, err
	}

	err = cert.Sign(ca.privKey, openssl.EVP_SHA256)
	if err != nil {
		return nil, err
	}

	return &openCert{
		privKey: privKey,
		cert:    cert,
	}, nil
}

func (ca *OpenCA) Issue(host string, vaildFor time.Duration, rsaBits int) (*tls.Certificate, error) {
	cert, err := ca.issue(host, vaildFor, rsaBits)
	if err != nil {
		return nil, err
	}

	certBytes, err := cert.cert.MarshalPEM()
	if err != nil {
		return nil, err
	}

	privKeyBytes, err := cert.privKey.MarshalPKCS1PrivateKeyPEM()
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair(certBytes, privKeyBytes)
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}

func (ca *OpenCA) IssueFile(host string, vaildFor time.Duration, rsaBits int) (string, error) {
	cert, err := ca.issue(host, vaildFor, rsaBits)
	if err != nil {
		return "", err
	}

	commonname, err := getCommonName(host)
	if err != nil {
		return "", err
	}
	filename := commonname + ".crt"

	outFile, err := os.Create(filename)
	defer outFile.Close()
	if err != nil {
		return "", err
	}

	certBytes, err := cert.cert.MarshalPEM()
	if err != nil {
		return "", err
	}
	_, err = outFile.Write(certBytes)
	if err != nil {
		return "", err
	}

	privKeyBytes, err := cert.privKey.MarshalPKCS1PrivateKeyPEM()
	if err != nil {
		return "", err
	}
	_, err = outFile.Write(privKeyBytes)
	if err != nil {
		return "", err
	}
	return filename, nil
}
