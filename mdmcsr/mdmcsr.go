// Package mdmcsr provides utilites for generating Apple "MDM CSR" Push Certificate signature requests.
package mdmcsr

import (
	"crypto"
	"encoding/base64"
	"encoding/pem"
	"io"

	"github.com/micromdm/plist"
)

// PushCertRequest represents the "MDM CSR"-signed plist structure
// required by https://identity.apple.com to get an Apple-signed MDM Push certificate.
// See https://developer.apple.com/documentation/devicemanagement/implementing_device_management/setting_up_push_notifications_for_your_mdm_customers
type PushCertRequest struct {
	PushCertRequestCSR       string `plist:"PushCertRequestCSR"`
	PushCertCertificateChain string `plist:"PushCertCertificateChain"`
	PushCertSignature        string `plist:"PushCertSignature"`
}

// New creates a new push certificate signed request.
func New() *PushCertRequest {
	return new(PushCertRequest)
}

// AppendCertificateChain encodes certs into PEM blocks and appends them to p.
// The certs are expected to be DER-encoded.
func (r *PushCertRequest) AppendCertificateChain(certs ...[]byte) {
	if len(certs) < 1 {
		return
	}
	block := &pem.Block{Type: "CERTIFICATE"}
	for _, cert := range certs {
		block.Bytes = cert
		r.PushCertCertificateChain += string(pem.EncodeToMemory(block))
	}
}

// Sign creates a new signed push certificate request by signing csr
// with signer using the certificate chain from template.
func Sign(rand io.Reader, template *PushCertRequest, csr []byte, hashType crypto.Hash, signer crypto.Signer) (*PushCertRequest, error) {
	h := hashType.New()
	if _, err := h.Write(csr); err != nil {
		return nil, err
	}
	sig, err := signer.Sign(rand, h.Sum(nil), hashType)
	if err != nil {
		return nil, err
	}
	r := &PushCertRequest{
		PushCertRequestCSR: base64.StdEncoding.EncodeToString(csr),
		PushCertSignature:  base64.StdEncoding.EncodeToString(sig),
	}
	if template != nil {
		r.PushCertCertificateChain = template.PushCertCertificateChain
	}
	return r, nil
}

// Encode marshals p into a Base64-encoded plist.
// This is the form of the request that https://identity.apple.com requires.
func (p *PushCertRequest) Encode() (string, error) {
	plist, err := plist.MarshalIndent(p, "\t")
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(plist), nil
}
