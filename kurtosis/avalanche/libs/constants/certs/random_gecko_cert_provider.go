// (c) 2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"time"

	mathrand "math/rand"

	"github.com/palantir/stacktrace"
)

const (
	certificatePreamble = "CERTIFICATE"
	privateKeyPreamble  = "RSA PRIVATE KEY"
)

var rootCert = x509.Certificate{
	SerialNumber: big.NewInt(2020),
	Subject: pkix.Name{
		Organization:  []string{"Kurtosis Technologies"},
		Country:       []string{"USA"},
		Province:      []string{""},
		Locality:      []string{""},
		StreetAddress: []string{""},
		PostalCode:    []string{""},
	},
	NotBefore:             time.Now(),
	NotAfter:              time.Now().AddDate(10, 0, 0),
	IsCA:                  true,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	BasicConstraintsValid: true,
}

// RandomAvalancheCertProvider implements AvalancheCertProviders by providing certs signed by the same root CA
type RandomAvalancheCertProvider struct {
	nextSerialNumber int64
	varyCerts        bool
}

// NewRandomAvalancheCertProvider creates a new cert provider that can optionally return either the same cert every time, or different ones
// Args:
// 	varyCerts: True to produce a different cert on each call to GetCertAndKey, or false to yield the same
// 		randomly-generated cert each time
func NewRandomAvalancheCertProvider(varyCerts bool) *RandomAvalancheCertProvider {
	return &RandomAvalancheCertProvider{
		nextSerialNumber: mathrand.Int63(), // nolint:gosec
		varyCerts:        varyCerts,
	}
}

// GetCertAndKey implements AvalancheCertProvider function that yields a new cert and private key based off the configuration parameters
// defined at construction time
// Returns:
// 	certPemBytes: The bytes of the generated cert
// 	keyPemBytes: The bytes of the private key that was generated alongside the cert
func (r *RandomAvalancheCertProvider) GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error) {
	serialNum := r.nextSerialNumber
	if r.varyCerts {
		r.nextSerialNumber = mathrand.Int63() // nolint:gosec
	}
	serviceCert := getServiceCert(serialNum)

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, stacktrace.Propagate(err, "Failed to generate random private key.")
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, serviceCert, &rootCert, &(certPrivKey.PublicKey), certPrivKey)
	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, stacktrace.Propagate(err, "Failed to sign service cert with cert authority.")
	}
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  certificatePreamble,
		Bytes: certBytes,
	}); err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  privateKeyPreamble,
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	}); err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	return *certPEM, *certPrivKeyPEM, nil
}

// ================= Helper functions ===================
func getServiceCert(serialNumber int64) *x509.Certificate {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(serialNumber),
		Subject: pkix.Name{
			Organization:  []string{"Kurtosis Technologies"},
			Country:       []string{"USA"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	return cert
}
