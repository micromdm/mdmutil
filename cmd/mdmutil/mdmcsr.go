package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"slices"
	"time"

	"github.com/micromdm/mdmutil/mdmcsr"
)

const (
	AppleRootCAURL       = "https://www.apple.com/appleca/AppleIncRootCertificate.cer"
	AppleIntermediateURL = "https://www.apple.com/certificateauthority/AppleWWDRCAG3.cer"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 5 * time.Second,
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
	},
	Timeout: 10 * time.Second,
}

func mdmcsrSign(name string, args []string, usage func()) int {
	f := flag.NewFlagSet(name, flag.ExitOnError)
	var (
		flRootCA        = f.String("root-ca", path.Base(AppleRootCAURL), "Path to read and save Apple Root CA in DER format; downloaded once from "+AppleRootCAURL)
		flIntermed      = f.String("intermediate", path.Base(AppleIntermediateURL), "Path to read and save Apple Intermediate Certificate in DER format; downloaded once from "+AppleIntermediateURL)
		flAPNsCSR       = f.String("apns-csr", "", "Path to APNs CSR that the MDMCSR private key will sign in PEM format")
		flMDMCSRPrivKey = f.String("mdmcsr-private-key", "", "Path to MDM CSR private key in PEM PKCS#1 or PKCS#8 format")
		flMDMCSRCert    = f.String("mdmcsr-certificate", "", "Path to MDM CSR certificate in DER or PEM format")
		flOut           = f.String("out", "-", "Output filename of the signed MDM CSR request; \"-\" for stdout")
		flVerbose       = f.Bool("v", false, "Print verbose messages to stderr")
	)
	cmdUsage(f, usage, nil, "")

	if err := f.Parse(args); err != nil {
		flagUsageExit(f, "failed to parse args", 2)
	}

	if *flAPNsCSR == "" || *flMDMCSRCert == "" || *flMDMCSRPrivKey == "" {
		flagUsageExit(f, "path to APNs CSR, MDM CSR cert, and MDM CSR priv key must all be specified", 2)
	}

	// download the root & intermediate certs if they don't exist
	for filename, url := range map[string]string{
		*flRootCA:   AppleRootCAURL,
		*flIntermed: AppleIntermediateURL,
	} {
		if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			if *flVerbose {
				fmt.Fprintf(os.Stderr, "downloading %s to %s\n", url, filename)
			}
			if err = downloadToFile(httpClient, url, filename); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		}
	}

	// then load the root & intermediate certs
	var certs [][]byte
	for _, filename := range []string{*flRootCA, *flIntermed} {
		if *flVerbose {
			fmt.Fprintf(os.Stderr, "reading certificate %s\n", filename)
		}
		cer, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		certs = append(certs, cer)
	}

	// load the MDM CSR certificate
	if *flVerbose {
		fmt.Fprintf(os.Stderr, "reading certificate %s\n", *flMDMCSRCert)
	}
	mdmcsrCertBytes, err := os.ReadFile(*flMDMCSRCert)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var pemBlock *pem.Block

	// parse the MDM CSR certificate as a sanity check
	if bytes.HasPrefix(mdmcsrCertBytes, []byte("-----BEGIN ")) {
		// if the certificate was encapsulated as PEM ...
		pemBlock, err = decodePEM(mdmcsrCertBytes, []string{"CERTIFICATE"})
		if err != nil {
			err = fmt.Errorf("mdmcsr certificate %s: decode PEM: %w", *flMDMCSRCert, err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		mdmcsrCertBytes = pemBlock.Bytes
	}

	// ... but usally the certificate comes DER-encoded from Apple
	_, err = x509.ParseCertificate(mdmcsrCertBytes)
	if err != nil {
		err = fmt.Errorf("mdmcsr certificate %s: parse certificate: %w", *flMDMCSRCert, err)
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	certs = append(certs, mdmcsrCertBytes)

	// load the CSR
	if *flVerbose {
		fmt.Fprintf(os.Stderr, "reading APNs CSR %s\n", *flAPNsCSR)
	}
	csrBytes, err := os.ReadFile(*flAPNsCSR)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// parse the CSR as a sanity check
	pemBlock, err = decodePEM(csrBytes, []string{"CERTIFICATE REQUEST"})
	if err != nil {
		err = fmt.Errorf("mdmcsr certificate %s: decode PEM: %w", *flAPNsCSR, err)
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	csr := pemBlock.Bytes
	_, err = x509.ParseCertificateRequest(pemBlock.Bytes)
	if err != nil {
		err = fmt.Errorf("apns csr %s: parse certificate request: %w", *flAPNsCSR, err)
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// load the MDM CSR private key
	if *flVerbose {
		fmt.Fprintf(os.Stderr, "reading MDM CSR private key %s\n", *flMDMCSRPrivKey)
	}
	privKeyBytes, err := os.ReadFile(*flMDMCSRPrivKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// parse the MDM CSR private key for signing usage
	pemBlock, err = decodePEM(privKeyBytes, []string{"RSA PRIVATE KEY", "PRIVATE KEY"})
	if err != nil {
		err = fmt.Errorf("mdmcsr private key %s: decode PEM: %w", *flMDMCSRPrivKey, err)
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var signer crypto.Signer
	switch pemBlock.Type {
	case "RSA PRIVATE KEY":
		if procType, ok := pemBlock.Headers["Proc-Type"]; ok {
			err = fmt.Errorf("encrypted PEM not supported: Proc-Type=%s", procType)
		} else {
			signer, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		}
	case "PRIVATE KEY":
		p8key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err == nil {
			var ok bool
			signer, ok = p8key.(crypto.Signer)
			if !ok {
				err = fmt.Errorf("key is not a crypto signer")
			}
		}
	default:
		err = fmt.Errorf("unknown PEM block type: %s", pemBlock.Type)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("mdmcsr private key %s: %w", *flMDMCSRPrivKey, err))
		return 1
	}

	// instantiate our MDM CSR template
	mdmcsrTemplate := mdmcsr.New()
	mdmcsrTemplate.AppendCertificateChain(certs...)

	// sign the CSR
	if *flVerbose {
		fmt.Fprintln(os.Stderr, "signing")
	}
	mdmcsrSigned, err := mdmcsr.Sign(rand.Reader, mdmcsrTemplate, csr, crypto.SHA256, signer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// encode it in the identity.apple.com format
	encoded, err := mdmcsrSigned.Encode()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// finally, output our result
	var out io.StringWriter
	if *flOut == "-" {
		out = os.Stdout
	} else {
		if *flVerbose {
			fmt.Fprintf(os.Stderr, "writing %s\n", *flOut)
		}
		f, err := os.Create(*flOut)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer f.Close()
		out = f
	}
	_, err = out.WriteString(encoded)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

type HTTPGetter interface {
	Get(url string) (resp *http.Response, err error)
}

func downloadToFile(getter HTTPGetter, url, filepath string) error {
	if getter == nil {
		getter = http.DefaultClient
	}

	resp, err := getter.Get(url)
	if err != nil {
		return fmt.Errorf("failed get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status fetching %s: %s", url, resp.Status)
	}

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to download body: %w", err)
	}

	return nil
}

func decodePEM(in []byte, allowedTypes []string) (*pem.Block, error) {
	b, _ := pem.Decode(in)
	if b == nil || len(b.Bytes) <= 0 {
		return nil, errors.New("empty PEM block")
	}
	if len(allowedTypes) > 0 && !slices.Contains(allowedTypes, b.Type) {
		return b, fmt.Errorf("invalid PEM block type: %s", b.Type)
	}
	return b, nil
}
