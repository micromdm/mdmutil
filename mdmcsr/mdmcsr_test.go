package mdmcsr

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	_ "crypto/sha1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/micromdm/plist"
)

func decodeCertChain(p *PushCertRequest) (certs [][]byte, err error) {
	rest := []byte(p.PushCertCertificateChain)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			err = errors.New("failed to decode certificate PEM block")
			return
		}
		certs = append(certs, block.Bytes)
	}
	return
}

func TestSign(t *testing.T) {
	// test data
	crt1Bytes := []byte{1, 2, 3}
	crt2Bytes := []byte{4, 5, 6}
	csr := []byte{7, 8, 9}

	// make a new key
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	// make new request
	tmpl := New()

	// append our (fake) certs
	tmpl.AppendCertificateChain(crt1Bytes, crt2Bytes)

	// sign and encode
	signed, err := Sign(rand.Reader, tmpl, csr, crypto.SHA1, key)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := signed.Encode()
	if err != nil {
		t.Fatal(err)
	}

	// decode our newly encoded request
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	recv := New()
	err = plist.Unmarshal(decoded, recv)
	if err != nil {
		t.Fatal(err)
	}
	certs, err := decodeCertChain(recv)
	if err != nil {
		t.Fatal(err)
	}
	if len(certs) != 2 {
		t.Fatal("incorrect number of certs")
	}

	// verify
	if !bytes.Equal(certs[0], crt1Bytes) {
		t.Error("cert 1 does not match")
	}
	if bytes.Compare(certs[1], crt1Bytes) != 1 {
		t.Error("cert 2 does not match")
	}
	recvCSR, err := base64.StdEncoding.DecodeString(recv.PushCertRequestCSR)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(recvCSR, csr) {
		t.Error("csr does not match")
	}
	sig, err := base64.StdEncoding.DecodeString(recv.PushCertSignature)
	if err != nil {
		t.Fatal(err)
	}
	h := crypto.SHA1.New()
	h.Write(recvCSR)
	err = rsa.VerifyPKCS1v15(&key.PublicKey, crypto.SHA1, h.Sum(nil), sig)
	if err != nil {
		t.Fatal(err)
	}
}
