package certs

import "bytes"

// StaticAvalancheCertProvider implements AvalancheCertProvider and provides the same cert every time
type StaticAvalancheCertProvider struct {
	key  bytes.Buffer
	cert bytes.Buffer
}

// NewStaticAvalancheCertProvider creates an instance of StaticAvalancheCertProvider using the given key and cert
// Args:
// 	key: The private key that the StaticAvalancheCertProvider will return on every call to GetCertAndKey
// 	cert: The cert that will be returned on every call to GetCertAndKey
func NewStaticAvalancheCertProvider(key bytes.Buffer, cert bytes.Buffer) *StaticAvalancheCertProvider {
	return &StaticAvalancheCertProvider{key: key, cert: cert}
}

// GetCertAndKey returns the same cert and key that was configured at the time of construction
func (s StaticAvalancheCertProvider) GetCertAndKey() (certPemBytes bytes.Buffer, keyPemBytes bytes.Buffer, err error) {
	return s.cert, s.key, nil
}
